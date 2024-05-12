# how to install

Make sure you have golang installed (https://golang.org/doc/install),
and then run

- `go install github.com/Merith-TK/packwiz-wrapper/cmd/pw@main`

# how to use

- `pw version`
  - prints pw version
- `pw help`
  - prints help
- `pw import <import.txt>`
  - imports mods from links in import.txt
  - URL files now support specifying which folder they go in (see #import)
- `pw modlist [raw] [versions]`
  - Generates a modlist file
  - arguments are keyword matches, so you can do `pw modlist raw versions` or `pw modlist versions raw`
  - You can use `raw` to generate without markdown formatting, (can be used for import)
  - You can use `versions` to generate with versions specified in modlist
- `pw reinstall`
  - reinstalls all meta-files, including URLs
- `pw batch (arguments)`
  - runs arguments in each subfolder of the current folder (changeable by using `-d <path>` before `batch`)
- `pw arb <arguments>`
  - runs arbitrary commands, useful for batchmode otherwise not really
## import

`pw import` now supports importing from a file, `import.txt` by default.

- this file takes an URL per line, and will import all mods from those URLs (where possible)
- currently supports importing from curseforge, modrinth, and URL files (URL files have formatting)

## modlist

`pw modlist` generates a modlist file, `modlist.md` by default.

- This is automatically sorted (where possible) for client, shared, and server mods,
- Shared mods are mods that are required on both client and server for full functionality
  - this is detected from the `side` field in the mod's `pack.toml`

By default, this is what the output will be:

```markdown
# Modlist

## Client Mods

- [Sodium](https://modrinth.com/mod/AANobbMI)

## Shared Mods

- [Lithium (Fabric)](https://www.curseforge.com/minecraft/mc-mods/lithium)
```

if you pass `raw` as an argument, it will generate without markdown formatting on the URL's to make it compatible with `pw import` like so:

```markdown
## Client Mods

Sodium
https://modrinth.com/mod/AANobbMI

## Shared Mods

Lithium (Fabric)
https://www.curseforge.com/minecraft/mc-mods/lithium
```

if you pass `versions` as an argument, it will generate with versions specified in modlist like so:

```markdown
# Modlist

## Client Mods

- [Sodium](https://modrinth.com/mod/AANobbMI/versions/b4hTi3mo)

## Shared Mods

- [Lithium (Fabric)](https://www.curseforge.com/minecraft/mc-mods/lithium/files/4439705)
```

Do note that this does not support URL files, and will not generate versions for those.

## reinstall

- `pw reinstall` will reinstall all meta-files, including URLs
  - It will load all meta-files into internal memory, and then reinstall them with packwiz.

## batch

- `pw batch` will run packwiz commands in all subfolders of a folder
  - run from the folder containing the folders with pack.toml's
  - `pw batch refresh` will refresh all subpacks
    ```
    pack1/
    pack2/
    ```
  - note, batchmode supports "recursion" such as `pw batch batch refresh` will run refresh in each sub/subfolder, 
    ```
    pack1/
      subpack1/
      subpack2/
    pack2/
      subpack1/
      subpack2/
    ```

## arb
- Runs arbitrary command from arguments
  - `pw arb mkdir .minecraft` will run `mkdir .minecraft`
  - really only useful if you are using batchmode as you can run a single arbitrary command in every pack folder

## flags

- `pw` supports a few flags, which can be used with any subcommand
  - `pw -h` will print help
  - `pw -r` will run `packwiz refresh` after operations
  - running `pw -r import` will automatically refresh after its done importing
  - `pw -y` will autoconfirm (not full implemented into all subcommands)
  - `pw -c` is depreciated
  - Originally used for importing only clientside mods, but does nothing now
  - `pw -d <PackDir>` will set the pack directory to `<PackDir>`
  - In batchmode this points to where folders containg pack.toml's are rather than the current folder

