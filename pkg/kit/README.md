# `pkg/kit` Go API

`pkg/kit` exposes reusable APIs for core Kit operations so Go programs can call Kit functionality directly without shelling out to the CLI.

## Example

```go
ctx := context.Background()

opts := &kit.ListOptions{
	ConfigHome: os.Getenv("KITOPS_HOME"),
}

kits, err := kit.List(ctx, opts)
if err != nil {
	log.Fatal(err)
}

for _, k := range kits {
	fmt.Printf("%s %v\n", k.Repo, k.Tags)
}
```

## Available operations

- `kit.Info`
- `kit.Inspect`
- `kit.List`
- `kit.Login`
- `kit.Logout`
- `kit.Pull`
- `kit.Push`
- `kit.Remove`
- `kit.Tag`
<!--  -->
