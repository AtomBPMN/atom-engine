/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package cli

import (
	"strings"

	"atom-engine/proto/expression/expressionpb"
)

// Helper function to format parameter list for display
func getParameterList(parameters []*expressionpb.ParameterInfo) string {
	if len(parameters) == 0 {
		return ""
	}

	paramNames := make([]string, len(parameters))
	for i, param := range parameters {
		paramNames[i] = param.Name
		if !param.Required {
			paramNames[i] = paramNames[i] + "?"
		}
	}

	return strings.Join(paramNames, ", ")
}
