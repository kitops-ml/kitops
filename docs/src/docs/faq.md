# KitOps FAQ

## Why does kit pack not support symlinks?

A symlink has no single correct meaning at pack time, so packing one would force Kit to guess your intent. Kit avoids the guess by skipping symbolic links during `pack`.

If you are using symlinks to reference shared data that lives elsewhere on disk, package that data as a separate ModelKit, then point to it with a remote reference in your primary ModelKit. The reference does the job the symlink was doing, and it stays valid when the ModelKit moves to another machine.

The reasons we skip symlinks are that both guesses on how to handle them create problems:

- Store the link as a link. The ModelKit keeps the symlink pointing at a path on your machine. That path will not exist when someone unpacks the ModelKit elsewhere, so the link is dead and the artifact is not portable. Links that point outside the pack context, such as `~/.ssh`, also raise a security concern, which leaves only the options of erroring or silently dropping them.
- Follow the link and pack its target. The unpacked ModelKit then contains a different set of files than the directory you packed, and it works against the usual reason for using a symlink, which is to avoid duplicating data. Ten symlinks to a 100GB file would produce a 1TB ModelKit.

## Why can't I pack things outside the context?

The context directory is the boundary of the ModelKit. Kit packs what is inside it and rejects paths that resolve outside it, including absolute paths and paths that climb out with `..`. Two reasons:

- Reproducibility. A ModelKit should unpack to the same set of files on any machine. If `pack` could reach arbitrary locations on the machine that built it, the contents would depend on that machine's layout and would not reproduce elsewhere.
- Security. Reaching outside the context would let a Kitfile read files anywhere the process has access, such as `~/.ssh` or `~/.aws/credentials`, and pull them into an artifact that gets pushed to a registry and shared. Keeping `pack` inside the context closes that path (although it's still your responsibility to ensure that there isn't sensitive data in the context).

If you need data that lives elsewhere, copy it into the context before packing, or package it as a separate ModelKit and reference it.
