package nvim

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/dradtke/debug-console/dap"
	"github.com/neovim/go-client/nvim"
	"github.com/neovim/go-client/nvim/plugin"
)

const (
	SignGroupBreakpoint = "debug-console-breakpoint"
	SignNameBreakpoint  = "debug-console-breakpoint"
)

var (
	dapDir string
)

func SendConfiguration(v *nvim.Nvim, p *dap.Process) error {
	log.Print("Sending configuration")

	allBreakpointSigns, err := GetAllSigns(v, SignGroupBreakpoint)
	if err != nil {
		return fmt.Errorf("Error getting breakpoint signs: %w", err)
	}

	var (
		wg     sync.WaitGroup
		errs   []error
		errsMu sync.Mutex
	)

	addErr := func(err error) {
		errsMu.Lock()
		errs = append(errs, err)
		errsMu.Unlock()
	}

	wg.Add(len(allBreakpointSigns))

	for buffer, breakpointSigns := range allBreakpointSigns {
		go func(buffer nvim.Buffer, breakpointSigns []SignInfo) {
			bufferPath, err := BufferPath(v, buffer)
			if err != nil {
				addErr(fmt.Errorf("SendConfiguration: %w", err))
			}
			var breakpoints []map[string]any
			for _, breakpointSign := range breakpointSigns {
				breakpoints = append(breakpoints, map[string]any{
					"line": breakpointSign.LineNumber,
				})
			}
			if _, err := p.SendRequest("setBreakpoints", map[string]any{
				"source": map[string]any{
					"path": bufferPath,
				},
				"breakpoints": breakpoints,
			}); err != nil {
				addErr(fmt.Errorf("Error setting breakpoints: %w", err))
			}
			wg.Done()
		}(buffer, breakpointSigns)
	}

	log.Print("Waiting for breakpoint setting to complete")
	wg.Wait()

	// TODO: use multierr or similar?
	if len(errs) > 0 {
		log.Printf("Got an error: %s", errs[0])
		return fmt.Errorf("Error setting one or more breakpoints: %w", errs[0])
	}

	if _, err := p.SendRequest("configurationDone", make(map[string]any)); err != nil {
		return fmt.Errorf("Error finishing configuration: %w", err)
	}

	log.Print("Configuration complete!")
	return nil
}

func setLogOutput() error {
	filename := os.Getenv("LOG_FILE")
	if filename == "" {
		return nil
	}
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("setLogOutput: error opening log file: %w", err)
	}
	log.SetFlags(0)
	log.SetOutput(f)
	return nil
}

func HandlePanic() {
	if r := recover(); r != nil {
		log.Printf("Panic error: %s", r)
	}
}

func Main() error {
	if os.Getenv("NVIM") != "" {
		if err := setLogOutput(); err != nil {
			log.Print(err)
		}

		dapDir = os.Getenv("DAP_DIR")
		if _, err := os.Stat(dapDir); os.IsNotExist(err) {
			if err = os.MkdirAll(dapDir, 0644); err != nil {
				return fmt.Errorf("failed to create dap cache dir: %w", err)
			}
		}
	}

	plugin.Main(func(p *plugin.Plugin) error {
		d := &dap.DAP{
			Dir: os.Getenv("DAP_DIR"),
		}
		d.EventHandler = HandleEvent(p.Nvim, d) // this feels weird to do

		RegisterCommands(p, d)
		return nil
	})

	return nil
}
