# swarpf

-- This project is not production ready and might never be. Use at your own risk. --

This is a first iteration of a proxy framework designed for the mobile mobile game "Summoners War" by Com2Us.

The main component of the framework is an extensible proxy that can publish publish events to registered handlers over RPC.
There are example implementations of plugins in `cmd/plugins/`.

There is currently no focus on secure multi-user capability and at the moment there are no plans to implement such. Please only use this as single-user framework.