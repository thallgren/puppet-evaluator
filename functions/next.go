package functions

import (
	. "github.com/puppetlabs/go-evaluator/evaluator"
	. "github.com/puppetlabs/go-evaluator/errors"
)

func init() {
	NewGoFunction(`next`,
		func(d Dispatch) {
			d.OptionalParam(`Any`)
			d.Function(func(c EvalContext, args []PValue) PValue {
				arg := UNDEF
				if len(args) > 0 { arg = args[0] }
				panic(NewNextIteration(c.StackTop(), arg))
			})
		},
	)
}
