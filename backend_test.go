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
