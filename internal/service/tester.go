package service

import (
	"context"
	"fmt"
	"iter"
	"runtime"
	"sync"

	"github.com/kapetndev/tftest/internal/check"
)

// Finder defines an interface for finding Terraform modules.
type Finder interface {
	// FindModules returns a list of module paths to test.
	FindModules(ctx context.Context, rootPath string) ([]string, error)
}

// TestRequest represents a request to run tests on Terraform modules.
type TestRequest struct {
	RootPath string
	Verbose  bool
}

// TestService provides functionality to run tests on Terraform modules.
type TestService struct {
	checker      *check.Executor
	moduleFinder Finder
}

// NewTestService creates a new [TestService].
func NewTestService(checker *check.Executor, moduleFinder Finder) *TestService {
	return &TestService{
		checker:      checker,
		moduleFinder: moduleFinder,
	}
}

// ResultsIterator defines an iterator over slices of test results.
type ResultsIterator iter.Seq2[[]check.Result, error]

// Results returns an iterator over test results for each Terraform module.
func (s *TestService) Results(ctx context.Context, req TestRequest) ResultsIterator {
	return func(yield func([]check.Result, error) bool) {
		resultsCh, err := s.streamResults(ctx, req)
		if err != nil {
			yield(nil, err)
			return
		}

		// Stream results as they become available.
		for {
			select {
			case <-ctx.Done():
				yield(nil, ctx.Err())
				return
			case results, ok := <-resultsCh:
				if !ok {
					return
				}
				if !yield(results, nil) {
					// Caller stopped iterating - context cancellation will clean up the
					// goroutines.
					return
				}
			}
		}
	}
}

// CollectResults runs the tests on the specified Terraform modules.
func (s *TestService) CollectResults(ctx context.Context, req TestRequest) ([]check.Result, error) {
	var allResults []check.Result
	for results, err := range s.Results(ctx, req) {
		if err != nil {
			return nil, err
		}
		allResults = append(allResults, results...)
	}
	return allResults, nil
}

// streamResults streams test results for each Terraform module as they complete.
func (s *TestService) streamResults(ctx context.Context, req TestRequest) (<-chan []check.Result, error) {
	resultsCh := make(chan []check.Result, runtime.GOMAXPROCS(0))

	modules, err := s.moduleFinder.FindModules(ctx, req.RootPath)
	if err != nil {
		close(resultsCh)
		return resultsCh, fmt.Errorf("failed to find modules: %w", err)
	}

	var wg sync.WaitGroup
	wg.Add(len(modules))

	// Use a once-close approach for channels.
	closeOnce := sync.Once{}

	// Collector closes channels once all workers finish.
	go func() {
		wg.Wait()
		closeOnce.Do(func() {
			close(resultsCh)
		})
	}()

	// Semaphore to limit concurrency to number of CPU cores.
	sem := make(chan struct{}, runtime.GOMAXPROCS(0))

	for _, modulePath := range modules {
		go func(modulePath string) {
			defer wg.Done()

			// Acquire semaphore or exit if context is cancelled.
			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				return
			}
			defer func() { <-sem }()

			// Run all checks for this module.
			results := s.checker.RunAll(ctx, modulePath, check.RunOptions{Verbose: req.Verbose})

			// Stream results for this module.
			select {
			case resultsCh <- results:
			case <-ctx.Done():
				return
			}
		}(modulePath)
	}

	return resultsCh, nil
}
