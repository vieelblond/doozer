package paxos

import (
	"fmt"
)

const (
	iRnd = iota
	iNumParts
)

const (
	nRnd = iota
	nVal
	nNumParts
)

func acceptor(ins chan Msg, outs Putter) {
	var rnd, vrnd uint64
	var vval string

	update := func(in Msg) {
		defer swallowContinue()

		switch in.Cmd {
		case "INVITE":
			bodyParts := splitExactly(in.Body, iNumParts)

			i := dtoui64(bodyParts[iRnd])

			switch {
			case i <= rnd:
			case i > rnd:
				rnd = i

				reply := Msg{
					Seqn: in.Seqn,
					Cmd: "RSVP",
					To: in.From, // reply to the sender
					Body: fmt.Sprintf("%d:%d:%s", i, vrnd, vval),
				}
				outs.Put(reply)
			}
		case "NOMINATE":
			bodyParts := splitExactly(in.Body, nNumParts)

			i := dtoui64(bodyParts[nRnd])

			// SUPER IMPT MAD PAXOS
			if i < rnd || i == vrnd {
				return
			}

			rnd = i
			vrnd = i
			vval = bodyParts[nVal]

			broadcast := Msg{
				Seqn: in.Seqn,
				Cmd: "VOTE",
				To: 0,
				Body: fmt.Sprintf("%d:%s", i, vval),
			}
			outs.Put(broadcast)
		}
	}

	for in := range ins {
		update(in)
	}
}