// Copyright (c) OpenFaaS Author(s) 2019. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package types

import (
	"fmt"
	"log"
)

// ResponsePrinter prints function results
type ResponsePrinter struct {
	PrintResponseBody bool
}

// Response is triggered by the controller when a message is
// received from the function invocation
func (rp *ResponsePrinter) Response(res InvokerResponse) {
	if res.Error != nil {
		log.Printf("connector-sdk got error: %s", res.Error.Error())
	} else {
		log.Printf("connector-sdk got result: [%d] %s => %s (%d) bytes", res.Status, res.Topic, res.Function, len(res.Body))
		if rp.PrintResponseBody {
			fmt.Printf("[%d] %s => %s\n%s\n", res.Status, res.Topic, res.Function, string(res.Body))
		}
	}
}
