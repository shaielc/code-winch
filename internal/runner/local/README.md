# Local runner

The local runner exclusively owns sandbox execution handles and harness codecs.
It pumps attached bytes through the codec and publishes observations with a
monotonic, runner-local ordinal. It neither assigns canonical sequence numbers
nor persists events.
