package nvim

import (
	"fmt"
	"log"

	"github.com/neovim/go-client/nvim"
)

type SignGetPlacedResult struct {
	Buffer int `msgpack:"bufnr"` // NOTE: This won't marshal if using the nvim.Buffer type
	Signs  []struct {
		ID         int `msgpack:"id"`
		LineNumber int `msgpack:"lnum"`
	} `msgpack:"signs"`
}

type SignInfo struct {
	ID, LineNumber int
	Buffer         nvim.Buffer
	// BufferPattern is used if Buffer is 0
	BufferPattern string
	Group         string
	Exists        bool
}

func GetSignAt(v *nvim.Nvim, signGroup, buffer string, lineNum int) (SignInfo, error) {
	var result []SignGetPlacedResult
	if err := v.Call("sign_getplaced", &result, buffer, map[string]any{
		"group": signGroup,
		"lnum":  lineNum,
	}); err != nil {
		return SignInfo{}, fmt.Errorf("GetSignAt: Error getting sign info: %w", err)
	}

	sign := SignInfo{
		Buffer:     nvim.Buffer(result[0].Buffer),
		LineNumber: lineNum,
		Group:      signGroup,
	}

	if len(result[0].Signs) == 0 {
		return sign, nil
	}

	sign.Exists = true
	sign.ID = result[0].Signs[0].ID
	return sign, nil
}

func GetAllSigns(v *nvim.Nvim, signGroup string) (map[nvim.Buffer][]SignInfo, error) {
	signs := make(map[nvim.Buffer][]SignInfo)

	buffers, err := v.Buffers()
	if err != nil {
		log.Print(err)
		return signs, fmt.Errorf("GetAllSigns: %w", err)
	}

	for _, buffer := range buffers {
		var placedSigns []SignGetPlacedResult
		if err := v.Call("sign_getplaced", &placedSigns, buffer, map[string]any{
			"group": signGroup,
		}); err != nil {
			return signs, fmt.Errorf("GetAllSigns: Error getting sign info: %w", err)
		}
		for _, sign := range placedSigns[0].Signs {
			signs[buffer] = append(signs[buffer], SignInfo{
				Buffer:     buffer,
				ID:         sign.ID,
				LineNumber: sign.LineNumber,
				Group:      signGroup,
				Exists:     true,
			})
		}
	}

	return signs, nil
}

func BufferPath(v *nvim.Nvim, buffer nvim.Buffer) (string, error) {
	var result string
	if err := v.Call("expand", &result, fmt.Sprintf("#%d:p", buffer)); err != nil {
		return result, fmt.Errorf("BufferPath: %w", err)
	}
	return result, nil
}

func PlaceSign(v *nvim.Nvim, name string, sign SignInfo, priority int) error {
	var buffer any
	if sign.Buffer != 0 {
		buffer = sign.Buffer
	} else {
		buffer = sign.BufferPattern
	}
	if err := v.Call("sign_place", nil, 0, sign.Group, name, buffer, map[string]any{
		"lnum":     sign.LineNumber,
		"priority": priority,
	}); err != nil {
		return fmt.Errorf("PlaceSign: %w", err)
	}
	return nil
}

func RemoveSign(v *nvim.Nvim, sign SignInfo) error {
	if err := v.Call("sign_unplace", nil, sign.Group, map[string]any{
		"buffer": sign.Buffer,
		"id":     sign.ID,
	}); err != nil {
		return fmt.Errorf("RemoveSign: %w", err)
	}
	return nil
}

func RemoveAllSigns(v *nvim.Nvim, signGroup string) error {
	if err := v.Call("sign_unplace", nil, signGroup, map[string]any{}); err != nil {
		return fmt.Errorf("RemoveAllSigns: %w", err)
	}
	return nil
}

func GetLineNumber(v *nvim.Nvim) (int, error) {
	var lineNum int
	if err := v.Call("line", &lineNum, "."); err != nil {
		return 0, fmt.Errorf("GetLineNumber: %w", err)
	}
	return lineNum, nil
}
