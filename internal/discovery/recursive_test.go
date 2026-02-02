package discovery_test

import (
	"context"
	"errors"
	"testing"
	"testing/fstest"

	"github.com/google/go-cmp/cmp"

	"github.com/kapetndev/tftest/internal/discovery"
)

func TestRecursive_FindModules(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		fs       fstest.MapFS
		rootPath string
		maxDepth int
		want     []string
		wantErr  bool
	}{
		"single module in root": {
			fs: fstest.MapFS{
				"main.tf": &fstest.MapFile{Data: []byte("# terraform")},
			},
			rootPath: ".",
			want:     []string{"."},
		},
		"multiple modules in subdirectories": {
			fs: fstest.MapFS{
				"app/main.tf":     &fstest.MapFile{Data: []byte("# app module")},
				"db/main.tf":      &fstest.MapFile{Data: []byte("# db module")},
				"network/main.tf": &fstest.MapFile{Data: []byte("# network module")},
			},
			rootPath: ".",
			want:     []string{"app", "db", "network"},
		},
		"nested modules": {
			fs: fstest.MapFS{
				"services/api/main.tf": &fstest.MapFile{Data: []byte("# api module")},
				"services/web/main.tf": &fstest.MapFile{Data: []byte("# web module")},
				"infra/vpc/main.tf":    &fstest.MapFile{Data: []byte("# vpc module")},
			},
			rootPath: ".",
			want:     []string{"infra/vpc", "services/api", "services/web"},
		},
		"skips hidden directories": {
			fs: fstest.MapFS{
				"app/main.tf":           &fstest.MapFile{Data: []byte("# app")},
				".git/config":           &fstest.MapFile{Data: []byte("# git")},
				".hidden/main.tf":       &fstest.MapFile{Data: []byte("# hidden")},
				".terraform/main.tf":    &fstest.MapFile{Data: []byte("# terraform")},
				"node_modules/main.tf":  &fstest.MapFile{Data: []byte("# node")},
				".terragrunt-cache/foo": &fstest.MapFile{Data: []byte("# cache")},
				"vendor/main.tf":        &fstest.MapFile{Data: []byte("# vendor")},
			},
			rootPath: ".",
			want:     []string{"app"},
		},
		"stops at module boundary": {
			fs: fstest.MapFS{
				"parent/main.tf":       &fstest.MapFile{Data: []byte("# parent")},
				"parent/child/main.tf": &fstest.MapFile{Data: []byte("# child")},
			},
			rootPath: ".",
			want:     []string{"parent"},
		},
		"respects max depth": {
			fs: fstest.MapFS{
				"level1/main.tf":               &fstest.MapFile{Data: []byte("# level1")},
				"level1/level2/main.tf":        &fstest.MapFile{Data: []byte("# level2")},
				"level1/level2/level3/main.tf": &fstest.MapFile{Data: []byte("# level3")},
			},
			rootPath: ".",
			maxDepth: 2,
			want:     []string{"level1"},
		},
		"module with multiple tf files": {
			fs: fstest.MapFS{
				"app/main.tf":      &fstest.MapFile{Data: []byte("# main")},
				"app/variables.tf": &fstest.MapFile{Data: []byte("# vars")},
				"app/outputs.tf":   &fstest.MapFile{Data: []byte("# outputs")},
			},
			rootPath: ".",
			want:     []string{"app"},
		},
		"mixed tf and non-tf files": {
			fs: fstest.MapFS{
				"app/main.tf":   &fstest.MapFile{Data: []byte("# terraform")},
				"app/README.md": &fstest.MapFile{Data: []byte("# readme")},
				"app/script.sh": &fstest.MapFile{Data: []byte("#!/bin/bash")},
			},
			rootPath: ".",
			want:     []string{"app"},
		},
		"deep nesting with modules at various levels": {
			fs: fstest.MapFS{
				"a/main.tf":       &fstest.MapFile{Data: []byte("# a")},
				"a/b/c/main.tf":   &fstest.MapFile{Data: []byte("# c")},
				"a/b/d/e/main.tf": &fstest.MapFile{Data: []byte("# e")},
				"x/y/main.tf":     &fstest.MapFile{Data: []byte("# y")},
			},
			rootPath: ".",
			want:     []string{"a", "x/y"},
		},
		"module at root with subdirectories": {
			fs: fstest.MapFS{
				"main.tf":           &fstest.MapFile{Data: []byte("# root")},
				"subdir/other.txt":  &fstest.MapFile{Data: []byte("# should be ignored")},
				"subdir/nested.txt": &fstest.MapFile{Data: []byte("# also ignored")},
			},
			rootPath: ".",
			want:     []string{"."},
		},
		"only tfvars files should not be a module": {
			fs: fstest.MapFS{
				"app/terraform.tfvars": &fstest.MapFile{Data: []byte("# vars")},
				"app/dev.tfvars":       &fstest.MapFile{Data: []byte("# dev vars")},
			},
			rootPath: ".",
			want:     nil,
			wantErr:  true,
		},
		"tf extension in middle of filename": {
			fs: fstest.MapFS{
				"app/main.tf.backup": &fstest.MapFile{Data: []byte("# backup")},
				"app/test.txt":       &fstest.MapFile{Data: []byte("# text")},
			},
			rootPath: ".",
			want:     nil,
			wantErr:  true,
		},
		"no modules found": {
			fs: fstest.MapFS{
				"README.md":     &fstest.MapFile{Data: []byte("# readme")},
				"app/README.md": &fstest.MapFile{Data: []byte("# app readme")},
			},
			rootPath: ".",
			want:     nil,
			wantErr:  true,
		},
		"empty directory": {
			fs:       fstest.MapFS{},
			rootPath: ".",
			want:     nil,
			wantErr:  true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			finder := discovery.NewRecursiveModuleFinder(tt.fs)
			if tt.maxDepth > 0 {
				finder.MaxDepth = tt.maxDepth
			}

			got, err := finder.FindModules(context.Background(), tt.rootPath)
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error: %v, got: %v", tt.wantErr, err)
			}
			if !tt.wantErr {
				if diff := cmp.Diff(got, tt.want); diff != "" {
					t.Errorf("mismatch (-got +want):\n%s", diff)
				}
			}
		})
	}
}

func TestRecursive_FindModules_ContextCancellation(t *testing.T) {
	fs := fstest.MapFS{
		"a/main.tf": &fstest.MapFile{Data: []byte("# a")},
		"b/main.tf": &fstest.MapFile{Data: []byte("# b")},
		"c/main.tf": &fstest.MapFile{Data: []byte("# c")},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately.

	finder := discovery.NewRecursiveModuleFinder(fs)

	_, err := finder.FindModules(ctx, ".")
	if err == nil {
		t.Fatal("expected error due to cancelled context, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got: %v", err)
	}
}
