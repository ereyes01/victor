package victor

import (
    "github.com/brettbuddin/victor/utils/google"
)

func registerDefaultAbilities(brain *Brain) {
    brain.Respond("ping", func(ctx *Context) {
        ctx.Reply("pong!")
    })

    brain.Respond("(image|img) (.*)", func(ctx *Context) {
        result, err := google.ImageSearch(ctx.Matches()[3])

        if err != nil {
            return
        }

        if result == "" {
            ctx.Send("I didn't find anything.")
            return
        }

        ctx.Send(result)
    })

    brain.Respond("(list|show) users", func(ctx *Context) {
        list := ""

        for _, user := range brain.Users() {
            list += user.Name + "\n"
        }

        ctx.Paste(list)
    })
}