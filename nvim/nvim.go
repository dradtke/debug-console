package nvim

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/dradtke/debug-console/dap"
	"github.com/dradtke/debug-console/types"
	"github.com/dradtke/debug-console/util"
	"github.com/neovim/go-client/nvim"
	"github.com/neovim/go-client/nvim/plugin"
)

const (
	SignGroupBreakpoint = "debug-console-breakpoint"
	SignNameBreakpoint  = "debug-console-breakpoint"

	SignGroupCurrentLocation = "debug-console-current-location"
	SignNameCurrentLocation  = "debug-console-current-location"
)

var (
	dapDir string
)

func SendConfiguration(v *nvim.Nvim, p *dap.Conn) error {
	log.Print("Waiting for the initialized event...")
	<-p.InitializedEventSeen()

	log.Print("Setting breakpoints")

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
			defer util.Recover()
			bufferPath, err := BufferPath(v, buffer)
			if err != nil {
				addErr(fmt.Errorf("SendConfiguration: %w", err))
			}
			var breakpoints []types.SourceBreakpoint
			for _, breakpointSign := range breakpointSigns {
				breakpoints = append(breakpoints, types.SourceBreakpoint{
					Line: breakpointSign.LineNumber,
				})
			}
			if _, err := p.SendRequest(types.NewSetBreakpointRequest(types.SetBreakpointArguments{
				Source: types.Source{
					Path: &bufferPath,
				},
				Breakpoints: breakpoints,
			})); err != nil {
				addErr(fmt.Errorf("Error setting breakpoints: %w", err))
			}
			wg.Done()
		}(buffer, breakpointSigns)
	}

	wg.Wait()

	// TODO: use multierr or similar?
	if len(errs) > 0 {
		log.Printf("Got an error: %s", errs[0])
		return fmt.Errorf("Error setting one or more breakpoints: %w", errs[0])
	}

	// TODO: verify that the "supportsConfigurationDoneRequest" capability is true
	if _, err := p.ConfigurationDone(); err != nil {
		return fmt.Errorf("Error finishing configuration: %w", err)
	}

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

func Main(exe string) error {
	if os.Getenv("NVIM") != "" {
		if err := setLogOutput(); err != nil {
			log.Print(err)
		}
	}

	d := &dap.DAP{
		Exe: exe,
	}

	plugin.Main(func(p *plugin.Plugin) error {
		d.EditorEventHandler = HandleEvent(p.Nvim, d) // this feels weird to do
		p.HandleAutocmd(&plugin.AutocmdOptions{Event: "VimLeave", Pattern: "*"}, d.Stop)
		RegisterCommands(p, d)
		RegisterFunctions(p, d)
		return nil
	})

	return nil
}
