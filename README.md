[TempoDB's official Go package](http://github.com/tempodb/tempodb-go) is
powerful and flexible, but most of the time you just want to send data to it in
a simple fashion. This package is for that.

It has a very simple API: you just call `Send` for sending data. No need for
getting current time, creating `struct` instances, and checking for errors every
time.

It works asynchronously: Calling `Send` only puts the value in a channel and
doesn't block your code. Worker goroutines process that channel and and send the
data to TempoDB.

This package, at its current state, is useful for me and works fine in a
production app, but there are possible ways to improve it, anyways:

 - support for incemental calls for API
 - retrying failed requests
 - using bulk write and multi write APIs of TempoDB instead of sending data
   points one by one
