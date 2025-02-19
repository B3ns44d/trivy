package option

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	dbTypes "github.com/aquasecurity/trivy-db/pkg/types"
	"github.com/aquasecurity/trivy/pkg/types"
)

func TestReportReportConfig_Init(t *testing.T) {
	type fields struct {
		output         string
		Format         string
		Template       string
		vulnType       string
		securityChecks string
		severities     string
		IgnoreFile     string
		IgnoreUnfixed  bool
		listAllPksgs   bool
		ExitCode       int
		VulnType       []string
		Output         *os.File
		Severities     []dbTypes.Severity
		debug          bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    []string
		logs    []string
		want    ReportOption
		wantErr string
	}{
		{
			name: "happy path",
			fields: fields{
				severities:     "CRITICAL",
				vulnType:       "os",
				securityChecks: "vuln",
			},
			args: []string{"alpine:3.10"},
			want: ReportOption{
				Severities:     []dbTypes.Severity{dbTypes.SeverityCritical},
				VulnType:       []string{types.VulnTypeOS},
				SecurityChecks: []string{types.SecurityCheckVulnerability},
				Output:         os.Stdout,
			},
		},
		{
			name: "happy path with an unknown severity",
			fields: fields{
				severities:     "CRITICAL,INVALID",
				vulnType:       "os,library",
				securityChecks: "config",
			},
			args: []string{"centos:7"},
			logs: []string{
				"unknown severity option: unknown severity: INVALID",
			},
			want: ReportOption{
				Severities:     []dbTypes.Severity{dbTypes.SeverityCritical, dbTypes.SeverityUnknown},
				VulnType:       []string{types.VulnTypeOS, types.VulnTypeLibrary},
				SecurityChecks: []string{types.SecurityCheckConfig},
				Output:         os.Stdout,
			},
		},
		{
			name: "happy path with an cyclonedx",
			fields: fields{
				severities:     "CRITICAL",
				vulnType:       "os,library",
				securityChecks: "vuln",
				Format:         "cyclonedx",
				listAllPksgs:   true,
			},
			args: []string{"centos:7"},
			want: ReportOption{
				Severities:     []dbTypes.Severity{dbTypes.SeverityCritical},
				VulnType:       []string{types.VulnTypeOS, types.VulnTypeLibrary},
				SecurityChecks: []string{types.SecurityCheckVulnerability},
				Format:         "cyclonedx",
				Output:         os.Stdout,
				ListAllPkgs:    true,
			},
		},
		{
			name: "happy path with an cyclonedx option list-all-pkgs is false",
			fields: fields{
				severities:     "CRITICAL",
				vulnType:       "os,library",
				securityChecks: "vuln",
				Format:         "cyclonedx",
				listAllPksgs:   false,
				debug:          true,
			},
			args: []string{"centos:7"},
			logs: []string{
				"'cyclonedx', 'spdx', and 'spdx-json' automatically enables '--list-all-pkgs'.",
				"Severities: CRITICAL",
			},
			want: ReportOption{
				Severities:     []dbTypes.Severity{dbTypes.SeverityCritical},
				VulnType:       []string{types.VulnTypeOS, types.VulnTypeLibrary},
				SecurityChecks: []string{types.SecurityCheckVulnerability},
				Format:         "cyclonedx",
				Output:         os.Stdout,
				ListAllPkgs:    true,
			},
		},
		{
			name: "invalid option combination: --template enabled without --format",
			fields: fields{
				Template:       "@contrib/gitlab.tpl",
				severities:     "LOW",
				vulnType:       "os",
				securityChecks: "vuln",
			},
			args: []string{"gitlab/gitlab-ce:12.7.2-ce.0"},
			logs: []string{
				"'--template' is ignored because '--format template' is not specified. Use '--template' option with '--format template' option.",
			},
			want: ReportOption{
				Output:         os.Stdout,
				Severities:     []dbTypes.Severity{dbTypes.SeverityLow},
				Template:       "@contrib/gitlab.tpl",
				VulnType:       []string{types.VulnTypeOS},
				SecurityChecks: []string{types.SecurityCheckVulnerability},
			},
		},
		{
			name: "invalid option combination: --template and --format json",
			fields: fields{
				Format:         "json",
				Template:       "@contrib/gitlab.tpl",
				severities:     "LOW",
				vulnType:       "os",
				securityChecks: "config",
			},
			args: []string{"gitlab/gitlab-ce:12.7.2-ce.0"},
			logs: []string{
				"'--template' is ignored because '--format json' is specified. Use '--template' option with '--format template' option.",
			},
			want: ReportOption{
				Format:         "json",
				Output:         os.Stdout,
				Severities:     []dbTypes.Severity{dbTypes.SeverityLow},
				Template:       "@contrib/gitlab.tpl",
				VulnType:       []string{types.VulnTypeOS},
				SecurityChecks: []string{types.SecurityCheckConfig},
			},
		},
		{
			name: "invalid option combination: --format template without --template",
			fields: fields{
				Format:         "template",
				severities:     "LOW",
				vulnType:       "os",
				securityChecks: "vuln",
			},
			args: []string{"gitlab/gitlab-ce:12.7.2-ce.0"},
			logs: []string{
				"'--format template' is ignored because '--template' is not specified. Specify '--template' option when you use '--format template'.",
			},
			want: ReportOption{
				Format:         "template",
				Output:         os.Stdout,
				Severities:     []dbTypes.Severity{dbTypes.SeverityLow},
				VulnType:       []string{types.VulnTypeOS},
				SecurityChecks: []string{types.SecurityCheckVulnerability},
			},
		},
		{
			name: "invalid option combination: --list-all-pkgs with --format table",
			fields: fields{
				Format:         "table",
				severities:     "LOW",
				vulnType:       "os",
				securityChecks: "vuln",
				listAllPksgs:   true,
			},
			args: []string{"centos:7"},
			logs: []string{
				`"--list-all-pkgs" cannot be used with "--format table". Try "--format json" or other formats.`,
			},
			want: ReportOption{
				Format:         "table",
				Output:         os.Stdout,
				Severities:     []dbTypes.Severity{dbTypes.SeverityLow},
				VulnType:       []string{types.VulnTypeOS},
				SecurityChecks: []string{types.SecurityCheckVulnerability},
				ListAllPkgs:    true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := zap.InfoLevel
			if tt.fields.debug {
				level = zap.DebugLevel
			}

			core, obs := observer.New(level)
			logger := zap.New(core)

			set := flag.NewFlagSet("test", 0)
			_ = set.Parse(tt.args)

			c := &ReportOption{
				output:         tt.fields.output,
				Format:         tt.fields.Format,
				Template:       tt.fields.Template,
				vulnType:       tt.fields.vulnType,
				securityChecks: tt.fields.securityChecks,
				severities:     tt.fields.severities,
				IgnoreFile:     tt.fields.IgnoreFile,
				IgnoreUnfixed:  tt.fields.IgnoreUnfixed,
				ExitCode:       tt.fields.ExitCode,
				ListAllPkgs:    tt.fields.listAllPksgs,
				Output:         tt.fields.Output,
			}
			err := c.Init(os.Stdout, logger.Sugar())

			// tests log messages
			var gotMessages []string
			for _, entry := range obs.AllUntimed() {
				gotMessages = append(gotMessages, entry.Message)
			}
			assert.Equal(t, tt.logs, gotMessages, tt.name)

			// test the error
			switch {
			case tt.wantErr != "":
				require.NotNil(t, err)
				assert.Contains(t, err.Error(), tt.wantErr, tt.name)
				return
			}

			assert.NoError(t, err, tt.name)
			assert.Equal(t, &tt.want, c, tt.name)
		})
	}
}
