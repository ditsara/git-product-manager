# GPM — Git Product Manager

This project uses GPM for ticket management.

Do not create files or edit YAML front matter manually -- use the `pm` CLI.
Ticket, milestone, and comment markdown content can be edited directly.
After editing ticket content directly, run `pm edit <id> --touch` to update the timestamp.

Get started:
  pm --help                   # list all commands
  pm ai guide --help          # see available guide sections
  pm ai guide workflow        # read the development workflow
  pm list                     # show open tickets
  pm show <id>                # read a ticket
