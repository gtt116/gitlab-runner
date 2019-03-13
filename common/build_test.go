package common

import (
	"os"
	"testing"

	"errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func init() {
	s := MockShell{}
	s.On("GetName").Return("script-shell")
	s.On("GenerateScript", mock.Anything, mock.Anything).Return("script", nil)
	RegisterShell(&s)
}

func TestBuildRun(t *testing.T) {
	e := MockExecutor{}
	defer e.AssertExpectations(t)

	p := MockExecutorProvider{}
	defer p.AssertExpectations(t)

	// Create executor only once
	p.On("Create").Return(&e).Once()

	// We run everything once
	e.On("Prepare", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	e.On("Finish", nil).Return().Once()
	e.On("Cleanup").Return().Once()

	// Run script successfully
	e.On("Shell").Return(&ShellScriptInfo{Shell: "script-shell"})
	e.On("Run", mock.Anything).Return(nil)

	RegisterExecutor("build-run-test", &p)

	successfulBuild, err := GetSuccessfulBuild()
	assert.NoError(t, err)
	build := &Build{
		GetBuildResponse: successfulBuild,
		Runner: &RunnerConfig{
			RunnerSettings: RunnerSettings{
				Executor: "build-run-test",
			},
		},
	}
	err = build.Run(&Config{}, &Trace{Writer: os.Stdout})
	assert.NoError(t, err)
}

func TestRetryPrepare(t *testing.T) {
	PreparationRetryInterval = 0

	e := MockExecutor{}
	defer e.AssertExpectations(t)

	p := MockExecutorProvider{}
	defer p.AssertExpectations(t)

	// Create executor
	p.On("Create").Return(&e).Times(3)

	// Prepare plan
	e.On("Prepare", mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("prepare failed")).Twice()
	e.On("Prepare", mock.Anything, mock.Anything, mock.Anything).
		Return(nil).Once()
	e.On("Cleanup").Return().Times(3)

	// Succeed a build script
	e.On("Shell").Return(&ShellScriptInfo{Shell: "script-shell"})
	e.On("Run", mock.Anything).Return(nil)
	e.On("Finish", nil).Return().Once()

	RegisterExecutor("build-run-retry-prepare", &p)

	successfulBuild, err := GetSuccessfulBuild()
	assert.NoError(t, err)
	build := &Build{
		GetBuildResponse: successfulBuild,
		Runner: &RunnerConfig{
			RunnerSettings: RunnerSettings{
				Executor: "build-run-retry-prepare",
			},
		},
	}
	err = build.Run(&Config{}, &Trace{Writer: os.Stdout})
	assert.NoError(t, err)
}

func TestPrepareFailure(t *testing.T) {
	PreparationRetryInterval = 0

	e := MockExecutor{}
	defer e.AssertExpectations(t)

	p := MockExecutorProvider{}
	defer p.AssertExpectations(t)

	// Create executor
	p.On("Create").Return(&e).Times(3)

	// Prepare plan
	e.On("Prepare", mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("prepare failed")).Times(3)
	e.On("Cleanup").Return().Times(3)

	RegisterExecutor("build-run-prepare-failure", &p)

	successfulBuild, err := GetSuccessfulBuild()
	assert.NoError(t, err)
	build := &Build{
		GetBuildResponse: successfulBuild,
		Runner: &RunnerConfig{
			RunnerSettings: RunnerSettings{
				Executor: "build-run-prepare-failure",
			},
		},
	}
	err = build.Run(&Config{}, &Trace{Writer: os.Stdout})
	assert.EqualError(t, err, "prepare failed")
}

func TestPrepareFailureOnBuildError(t *testing.T) {
	e := MockExecutor{}
	defer e.AssertExpectations(t)

	p := MockExecutorProvider{}
	defer p.AssertExpectations(t)

	// Create executor
	p.On("Create").Return(&e).Times(1)

	// Prepare plan
	e.On("Prepare", mock.Anything, mock.Anything, mock.Anything).
		Return(&BuildError{}).Times(1)
	e.On("Cleanup").Return().Times(1)

	RegisterExecutor("build-run-prepare-failure-on-build-error", &p)

	successfulBuild, err := GetSuccessfulBuild()
	assert.NoError(t, err)
	build := &Build{
		GetBuildResponse: successfulBuild,
		Runner: &RunnerConfig{
			RunnerSettings: RunnerSettings{
				Executor: "build-run-prepare-failure-on-build-error",
			},
		},
	}
	err = build.Run(&Config{}, &Trace{Writer: os.Stdout})
	assert.IsType(t, err, &BuildError{})
}

func TestRunFailure(t *testing.T) {
	e := MockExecutor{}
	defer e.AssertExpectations(t)

	p := MockExecutorProvider{}
	defer p.AssertExpectations(t)

	// Create executor
	p.On("Create").Return(&e).Once()

	// Prepare plan
	e.On("Prepare", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	e.On("Cleanup").Return().Once()

	// Fail a build script
	e.On("Shell").Return(&ShellScriptInfo{Shell: "script-shell"})
	e.On("Run", mock.Anything).Return(errors.New("build fail"))
	e.On("Finish", errors.New("build fail")).Return().Once()

	RegisterExecutor("build-run-run-failure", &p)

	successfulBuild, err := GetSuccessfulBuild()
	assert.NoError(t, err)
	build := &Build{
		GetBuildResponse: successfulBuild,
		Runner: &RunnerConfig{
			RunnerSettings: RunnerSettings{
				Executor: "build-run-run-failure",
			},
		},
	}
	err = build.Run(&Config{}, &Trace{Writer: os.Stdout})
	assert.EqualError(t, err, "build fail")
}

func TestGetSourcesRunFailure(t *testing.T) {
	e := MockExecutor{}
	defer e.AssertExpectations(t)

	p := MockExecutorProvider{}
	defer p.AssertExpectations(t)

	// Create executor
	p.On("Create").Return(&e).Once()

	// Prepare plan
	e.On("Prepare", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	e.On("Cleanup").Return()

	// Fail a build script
	e.On("Shell").Return(&ShellScriptInfo{Shell: "script-shell"})
	e.On("Run", mock.Anything).Return(nil).Once()
	e.On("Run", mock.Anything).Return(errors.New("build fail")).Times(3)
	e.On("Finish", errors.New("build fail")).Return().Once()

	RegisterExecutor("build-get-sources-run-failure", &p)

	successfulBuild, err := GetSuccessfulBuild()
	assert.NoError(t, err)
	build := &Build{
		GetBuildResponse: successfulBuild,
		Runner: &RunnerConfig{
			RunnerSettings: RunnerSettings{
				Executor: "build-get-sources-run-failure",
			},
		},
	}

	build.Variables = append(build.Variables, BuildVariable{Key: "GET_SOURCES_ATTEMPTS", Value: "3"})
	err = build.Run(&Config{}, &Trace{Writer: os.Stdout})
	assert.EqualError(t, err, "build fail")
}

func TestArtifactDownloadRunFailure(t *testing.T) {
	e := MockExecutor{}
	defer e.AssertExpectations(t)

	p := MockExecutorProvider{}
	defer p.AssertExpectations(t)

	// Create executor
	p.On("Create").Return(&e).Once()

	// Prepare plan
	e.On("Prepare", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	e.On("Cleanup").Return()

	// Fail a build script
	e.On("Shell").Return(&ShellScriptInfo{Shell: "script-shell"})
	e.On("Run", mock.Anything).Return(nil).Times(2)
	e.On("Run", mock.Anything).Return(errors.New("build fail")).Times(3)
	e.On("Finish", errors.New("build fail")).Return().Once()

	RegisterExecutor("build-artifacts-run-failure", &p)

	successfulBuild, err := GetSuccessfulBuild()
	assert.NoError(t, err)
	build := &Build{
		GetBuildResponse: successfulBuild,
		Runner: &RunnerConfig{
			RunnerSettings: RunnerSettings{
				Executor: "build-artifacts-run-failure",
			},
		},
	}

	build.Variables = append(build.Variables, BuildVariable{Key: "ARTIFACT_DOWNLOAD_ATTEMPTS", Value: "3"})
	err = build.Run(&Config{}, &Trace{Writer: os.Stdout})
	assert.EqualError(t, err, "build fail")
}

func TestRestoreCacheRunFailure(t *testing.T) {
	e := MockExecutor{}
	defer e.AssertExpectations(t)

	p := MockExecutorProvider{}
	defer p.AssertExpectations(t)

	// Create executor
	p.On("Create").Return(&e).Once()

	// Prepare plan
	e.On("Prepare", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	e.On("Cleanup").Return()

	// Fail a build script
	e.On("Shell").Return(&ShellScriptInfo{Shell: "script-shell"})
	e.On("Run", mock.Anything).Return(nil).Times(3)
	e.On("Run", mock.Anything).Return(errors.New("build fail")).Times(3)
	e.On("Finish", errors.New("build fail")).Return().Once()

	RegisterExecutor("build-cache-run-failure", &p)

	successfulBuild, err := GetSuccessfulBuild()
	assert.NoError(t, err)
	build := &Build{
		GetBuildResponse: successfulBuild,
		Runner: &RunnerConfig{
			RunnerSettings: RunnerSettings{
				Executor: "build-cache-run-failure",
			},
		},
	}

	build.Variables = append(build.Variables, BuildVariable{Key: "RESTORE_CACHE_ATTEMPTS", Value: "3"})
	err = build.Run(&Config{}, &Trace{Writer: os.Stdout})
	assert.EqualError(t, err, "build fail")
}

func TestRunWrongAttempts(t *testing.T) {
	e := MockExecutor{}

	p := MockExecutorProvider{}
	defer p.AssertExpectations(t)

	// Create executor
	p.On("Create").Return(&e)

	// Prepare plan
	e.On("Prepare", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	e.On("Cleanup").Return()

	// Fail a build script
	e.On("Shell").Return(&ShellScriptInfo{Shell: "script-shell"})
	e.On("Run", mock.Anything).Return(nil).Once()
	e.On("Run", mock.Anything).Return(errors.New("Number of attempts out of the range [1, 10] for stage: get_sources"))
	e.On("Finish", errors.New("Number of attempts out of the range [1, 10] for stage: get_sources")).Return()

	RegisterExecutor("build-run-attempt-failure", &p)

	successfulBuild, err := GetSuccessfulBuild()
	assert.NoError(t, err)
	build := &Build{
		GetBuildResponse: successfulBuild,
		Runner: &RunnerConfig{
			RunnerSettings: RunnerSettings{
				Executor: "build-run-attempt-failure",
			},
		},
	}

	build.Variables = append(build.Variables, BuildVariable{Key: "GET_SOURCES_ATTEMPTS", Value: "0"})
	err = build.Run(&Config{}, &Trace{Writer: os.Stdout})
	assert.EqualError(t, err, "Number of attempts out of the range [1, 10] for stage: get_sources")
}

func TestRunSuccessOnSecondAttempt(t *testing.T) {
	e := MockExecutor{}
	p := MockExecutorProvider{}

	// Create executor only once
	p.On("Create").Return(&e).Once()

	// We run everything once
	e.On("Prepare", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	e.On("Finish", mock.Anything).Return().Twice()
	e.On("Cleanup").Return().Twice()

	// Run script successfully
	e.On("Shell").Return(&ShellScriptInfo{Shell: "script-shell"})

	e.On("Run", mock.Anything).Return(nil)
	e.On("Run", mock.Anything).Return(errors.New("build fail")).Once()
	e.On("Run", mock.Anything).Return(nil)

	RegisterExecutor("build-run-success-second-attempt", &p)

	successfulBuild, err := GetSuccessfulBuild()
	assert.NoError(t, err)
	build := &Build{
		GetBuildResponse: successfulBuild,
		Runner: &RunnerConfig{
			RunnerSettings: RunnerSettings{
				Executor: "build-run-success-second-attempt",
			},
		},
	}

	build.Variables = append(build.Variables, BuildVariable{Key: "GET_SOURCES_ATTEMPTS", Value: "3"})
	err = build.Run(&Config{}, &Trace{Writer: os.Stdout})
	assert.NoError(t, err)
}

func TestGetRemoteURL(t *testing.T) {
	testCases := []struct {
		runner RunnerSettings
		result string
	}{
		{
			runner: RunnerSettings{
				CloneURL: "http://test.local/",
			},
			result: "http://gitlab-ci-token:1234567@test.local/h5bp/html5-boilerplate.git",
		},
		{
			runner: RunnerSettings{
				CloneURL: "https://test.local",
			},
			result: "https://gitlab-ci-token:1234567@test.local/h5bp/html5-boilerplate.git",
		},
		{
			runner: RunnerSettings{},
			result: "http://fallback.url",
		},
	}

	for _, tc := range testCases {
		build := &Build{
			Runner: &RunnerConfig{
				RunnerSettings: tc.runner,
			},
			allVariables: JobVariables{
				JobVariable{Key: "CI_JOB_TOKEN", Value: "1234567"},
				JobVariable{Key: "CI_PROJECT_PATH", Value: "h5bp/html5-boilerplate"},
			},
			JobResponse: JobResponse{
				GitInfo: GitInfo{RepoURL: "http://fallback.url"},
			},
		}

		assert.Equal(t, tc.result, build.GetRemoteURL())
	}
}
