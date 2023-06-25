## Check Package

A check answers the question 'Is third party, networked dependency *X* available?'

A check is an arbitrary class that conforms to the *Check* interface, `check.go`.

A check attempts to establish a connection to a dependency and reports whether it
is currently avaiable, with the configuration provided to the check.