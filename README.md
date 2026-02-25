# ROSSLIB

> we will win because we are insane
> we will win because we are regarded

a better version of goodreads

recently I got an email that the "didn't finish shelf" is a feature "coming soon" to goodreads. this is a company with a hideous product and 428 employees. suck my nuts

## guidance

Technical notes are in docs/documentation

TODOs and planning are in docs/planning

## nephewbot

nephewbot is our favorite IC. Human brains should work in the docs, assume nephewbot will write the code.

`nephewbot/` contains a script that runs Claude Code's `/worker` skill on a cron (every 2 hours, 5 iterations). It works through the TODO list in `docs/TODO.md` autonomously. Logs go to `nephewbot/nephewbot.log`. for now this is just a cron on tristan's server. maybe it'll be better someday.