// Package utilization implements the Utilization spec, available at
// https://source.datanerd.us/agents/agent-specs/blob/master/Utilization.md
package utilization

import (
	"runtime"

	"github.com/newrelic/go-agent/internal/logger"
	"github.com/newrelic/go-agent/internal/sysinfo"
)

const metadataVersion = 1

// Config controls the behavior of utilization information capture.
type Config struct {
	DetectAWS    bool
	DetectDocker bool
}

// Data contains utilization system information.
type Data struct {
	MetadataVersion   int      `json:"metadata_version"`
	LogicalProcessors int      `json:"logical_processors"`
	RAMMib            *uint64  `json:"total_ram_mib"`
	Hostname          string   `json:"hostname"`
	Vendors           *vendors `json:"vendors,omitempty"`
}

var (
	sampleRAMMib = uint64(1024)
	// SampleData contains sample utilization data useful for testing.
	SampleData = Data{
		MetadataVersion:   metadataVersion,
		LogicalProcessors: 16,
		RAMMib:            &sampleRAMMib,
		Hostname:          "my-hostname",
	}
)

type vendor struct {
	ID   string `json:"id,omitempty"`
	Type string `json:"type,omitempty"`
	Zone string `json:"zone,omitempty"`
}

type vendors struct {
	AWS    *vendor `json:"aws,omitempty"`
	Docker *vendor `json:"docker,omitempty"`
}

// Gather gathers system utilization data.
func Gather(config Config, lg logger.Logger) *Data {
	uDat := Data{
		MetadataVersion:   metadataVersion,
		Vendors:           &vendors{},
		LogicalProcessors: runtime.NumCPU(),
	}

	if config.DetectDocker {
		id, err := sysinfo.DockerID()
		if err != nil &&
			err != sysinfo.ErrDockerUnsupported &&
			err != sysinfo.ErrDockerNotFound {
			lg.Warn("error gathering Docker information", map[string]interface{}{
				"error": err.Error(),
			})
		} else if id != "" {
			uDat.Vendors.Docker = &vendor{ID: id}
		}
	}

	if config.DetectAWS {
		aws, err := getAWS()
		if nil == err {
			uDat.Vendors.AWS = aws
		} else if isAWSValidationError(err) {
			lg.Warn("AWS validation error", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	if uDat.Vendors.AWS == nil && uDat.Vendors.Docker == nil {
		uDat.Vendors = nil
	}

	host, err := sysinfo.Hostname()
	if nil == err {
		uDat.Hostname = host
	} else {
		lg.Warn("error getting hostname", map[string]interface{}{
			"error": err.Error(),
		})
	}

	bts, err := sysinfo.PhysicalMemoryBytes()
	if nil == err {
		mib := sysinfo.BytesToMebibytes(bts)
		uDat.RAMMib = &mib
	} else {
		lg.Warn("error getting memory", map[string]interface{}{
			"error": err.Error(),
		})
	}

	return &uDat
}
