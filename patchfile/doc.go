/*
Package patchfile implements a robust patchfile format for directory trees.


Patchfile format

A Codechain patchfile is a UTF-8 encoded file split into newline separated
lines. It starts with the following line which defines the patchfile version:

  codechain patchfile version 1

The second line gives the tree hash (see tree package) of the directory tree
the patchfile applies to (example):

  treehash 5998c63aca42e471297c0fa353538a93d4d4cfafe9a672df6989e694188b4a92


The main body of the patch file encodes file deletions, file additions, and
file diffs.
A file deletion is encoded as follows (example):

  - f 927d2cae58bb53cdd087bb7178afeff9dab8ec1691cbd01aeccae62559da2791 gopher.png

The '-' denotes a deletion. The other three entries are the same as file
entries of tree lists (see tree package).

A file addition is encoded as follows (example):

  + f ad125cc5c1fb680be130908a0838ca2235db04285bcdd29e8e25087927e7dd0d hello.go

The '+' denotes an addition. The other three entries are the same as file
entries of tree lists (see tree package).

After an addition the actual patch must follow, either in "dmppatch" (for
UTF-8 files) or in "ascii85" format (for binary files).
The "dmppatch" file format looks like the following (example):

  dmppatch 2
  @@ -0,0 +1,78 @@
  +package main%0A%0Aimport (%0A%09%22fmt%22%0A)%0A%0Afunc main() %7B%0A%09fmt.Println(%22hello world!%22)%0A%7D%0A

The number after "dmppatch" denotes the number of lines following containing
the actual UTF-8 patch.

The "ascii85" file format looks like the following (example):

  ascii85 2
  +,^C)8Mp-E!!DW60b/e#'ElcGar]O1ZH.;>ZnWJO:iLd/`5G7uXPR`iQmq0B\]npD=)8AK4gPQFI-+W_
  >oidmeIj`.fgNufo<4MB5*&XfkqnCOo9\::*WQ0?z!!*#^!R=9-%KImW!!

The number after "ascii85" denotes the number of lines following containing
the actual binary encoding. "ascii85" patches are not real patches, but always
encode the entire binary file.

A file diff is encoded as follows (example):

  - f ad125cc5c1fb680be130908a0838ca2235db04285bcdd29e8e25087927e7dd0d hello.go
  + f 1b239e494fa201667627de82f0e4dc27b7b00b6ec06146e4d062730bf3762141 hello.go

As with file additions, after a file diff the actual patch must follow, either
in "dmppatch" or "ascii85" format (see above), if the file hash changed. That
is, if just the file mode changed and the file hash stayed the same no patch
must follow.

File diffs are only used if the file names ("hello.go" in the example above) are
the same. File moves are implemented as a file deletion and a file addition.

The last line in a patchfile must be the tree hash of the directory tree after
the patchfile has been applied (example):

  treehash e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855

Patchfiles are optimized for robustness, not for compactness or human
readability (although the human readability is reasonable).
A complete example containing a single UTF-8 file addition:

  codechain patchfile version 1
  treehash e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
  + f ad125cc5c1fb680be130908a0838ca2235db04285bcdd29e8e25087927e7dd0d hello.go
  dmppatch 2
  @@ -0,0 +1,78 @@
  +package main%0A%0Aimport (%0A%09%22fmt%22%0A)%0A%0Afunc main() %7B%0A%09fmt.Println(%22hello world!%22)%0A%7D%0A
  treehash 5998c63aca42e471297c0fa353538a93d4d4cfafe9a672df6989e694188b4a92


Diff function specification

Given the patchfile format described above, a Diff function that computes a
patch file for two directory trees rooted at A and B is straightforward to
implement:

  1. Calculate tree lists LIST_A and LIST_B (in lexical order) for A and B.

  2. Compare the file names NAME_A and NAME_B (lexicographically) of the first
     two entries in  LIST_A and LIST_B:

     - If NAME_A < NAME_B: File delete NAME_A, remove it from LIST_A, goto 2.
     - If NAME_A > NAME_B: File add NAME_B, remove it from LIST_B, goto 2.
     - If NAME_A == NAME_B:
       - If file mode or file hash of files NAME_A and NAME_B differ: file diff.
       - Remove NAME_A from LIST_A, NAME_B from LIST_B, and goto 2.

  3. If LIST_A still contains entries while LIST_B is empty, add file deletions
     for all entries in LIST_A.

  4. If LIST_B still contains entries while LIST_A is empty, add file additions
     for all entries in LIST_B.


Apply function specification

To apply a patchfile PATCH to a directory DIR we use the following algorithm:

  1. Read first line of of PATCH and make sure it contains a codechain
     patchfile version we understand.

  2. Read the second line of PATCH, make sure it is a treehash, and compare it
     with the treehash of DIR (before any patches have been applied).

  3. Read next line of PATCH:

     - If it starts with '+': Add file encoded in the following patch.
     - If it starts with '-':
       - If the next line starts with '+':
         - If the file name differ: Delete first file, add second file (with
           the following patch, which must be either ascii85 or dmppatch).
         - Otherwise (file names are the same):
           - If hashes are the same (only file modes differ): Adjust mode.
           - Otherwise (hashes differ): Apply the following patch, which must be
             either ascii85 or dmppatch (and adjust mode, if necessary).
       - Otherwise: Delete file.
     - If it starts with 'treehash': Goto 4.
     - Goto 3.

  4. Read the last line of PATCH, make sure it is a treehash, and compare it
     with the treehash of DIR (after all patches have been applied).

*/
package patchfile
