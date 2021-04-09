// Copyright (c) OpenFaaS Author(s) 2019. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package types

// ResponseSubscriber enables connector or another client in connector
// to receive results from the function invocation.
// Note: when implementing this interface, you must not perform any
// costly or high-latency operations, or should off-load them to another
// go-routine to prevent blocking.
type ResponseSubscriber interface {
	// Response is triggered by the controller when a message is
	// received from the function invocation
	Response(InvokerResponse)
}
