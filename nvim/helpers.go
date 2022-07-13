package nvim

import (
	"fmt"

	"github.com/neovim/go-client/nvim"
)

type SignGetPlacedResult struct {
	Buffer nvim.Buffer `msgpack:"bufnr"`
	Signs  []struct {
		ID         int `msgpack:"id"`
		LineNumber int `msgpack:"lnum"`
	} `msgpack:"signs"`
}

type SignInfo struct {
	ID, LineNumber int
	Buffer         nvim.Buffer
	Group          string
	Exists         bool
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
		Buffer:     result[0].Buffer,
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

func GetAllSigns(v *nvim.Nvim, signGroup string) ([]SignInfo, error) {
	var signs []SignInfo

	buffers, err := v.Buffers()
	if err != nil {
		return signs, fmt.Errorf("GetAllSigns: %w", err)
	}

	for _, buffer := range buffers {
		var result []SignGetPlacedResult
		if err := v.Call("sign_getplaced", &result, buffer, map[string]any{
			"group": signGroup,
		}); err != nil {
			return signs, fmt.Errorf("GetAllSigns: Error getting sign info: %w", err)
		}
		for _, sign := range result[0].Signs {
			signs = append(signs, SignInfo{
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

func PlaceSign(v *nvim.Nvim, name string, sign SignInfo) error {
	if err := v.Call("sign_place", nil, 0, sign.Group, name, sign.Buffer, map[string]any{
		"lnum":     sign.LineNumber,
		"priority": 99,
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

func GetLineNumber(v *nvim.Nvim) (int, error) {
	var lineNum int
	if err := v.Call("line", &lineNum, "."); err != nil {
		return 0, fmt.Errorf("GetLineNumber: %w", err)
	}
	return lineNum, nil
}
