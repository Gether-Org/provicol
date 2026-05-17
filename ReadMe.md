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


