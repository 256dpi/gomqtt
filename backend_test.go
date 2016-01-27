// Copyright (c) 2014 The gomqtt Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package broker

// store and look up retained messages
// look up retained messages with a # pattern
// look up retained messages with a + pattern
// remove retained message
// storing twice a retained message should keep only the last
// store and look up subscriptions by client
// remove subscriptions by client
// store and look up subscriptions by topic
// QoS 0 subscriptions, restored but not matched
// clean subscriptions
// store and count subscriptions
// add outgoing packet and stream it
// add outgoing packet and stream it twice
// add outgoing packet and update messageId
// add 2 outgoing packet and clear messageId
// update to pubrel
// add incoming packet, get it, and clear with messageId
// store, fetch and delete will message
// stream all will messages
// stream all will message for unknown brokers
