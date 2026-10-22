# Provicol

## How does it work

Provicol is a protocol question based. It's composed of a **parent** and a **child**.

The parent ask a question, the child answer, and the parent scan the answer.

The parent needs to connect first, here is the order of the connections:
```cpp
1. NewParent(path, perms) // give parent
2. NewChild(path) // give child

3. child.Bind(...) {
    child.Reply(...)
}
4. child.Listen()

5. parent.Ask(...).Scan(...)

6. parent.Close()
7. parent.Close()
```
(Thoses actions are grouped together because of what they do)

Everything in child needs to be implemented in the provider implementation, whild parent needs to be in the main software.

The child replies a buffer with a header

## Composition of buffer

Here is first what the header look like

<details>
    <summary> Previous Version </summary>
    ```
    {
        totalSize: uint64
    } (8 bytes)

    <data>
    ```
</details>

```
{
    magic:      uint32
    totalSize:  uint64
    isError:    uint32 // it could be bool but no.
} (16 bytes)

<data>
```

Magic is `0x1b505643` in hex, or this number in decimal `458249795`.

The isError is typically 0 or 1.

Ok, with that said, now let's see what each action does:

## `NewParent`

It takes a path and perms, basically, it creates a new unix socket that is bind to listen what the child will say

(note: the parent will remove the socket if it already exists)

It will wait until a new child connects, so be careful, it can freeze the program.

That's why I recommand to use `parent.NewParent` asynchrously (in a goroutine, or a thread) so the wainting doesnt affect the program.

## `NewChild`

It will connect to a socket given in parametter that a parent created. 

## File transfers

The Unix socket can also transfer an open file descriptor without putting the
file contents in a `[]byte`. This is useful for large local objects.

The parent can send a descriptor as the last handler argument:

```go
response := parent.AskWithFD(source, PutObject, publicKey, bucket, object)
if err := response.Scan(); err != nil {
    return err
}
```

The handler must accept `*os.File` as its last argument. `Listen` closes the
received descriptor when the handler returns, so a handler must not retain it:

```go
child.Bind(PutObject, func(publicKey, bucket, object string, source *os.File) error {
    _, err := io.Copy(destination, source)
    return err
})
```

For downloads, the child can attach a descriptor to its response:

```go
child.Bind(GetObject, func(publicKey, bucket, object string) error {
    file, err := os.Open(path)
    if err != nil {
        return err
    }
    return child.ReplyFD(file)
})

file, err := parent.Ask(GetObject, publicKey, bucket, object).ScanFD()
if err != nil {
    return err
}
defer file.Close()
```

`AskWithFD`, `ReplyFD` and `ScanFD` require the existing Unix socket
connection. File descriptors are local kernel resources and cannot be used
over TCP or between machines.
