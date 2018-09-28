Codechain walkthrough
---------------------

Start Codechain walkthrough with example project:

    $ cd doc/hellproject
    $ ls
    hello.go README.md

Let's generate a key pair for Alice:

    $ codechain keygen
    passphrase: 
    confirm passphrase: 
    comment (e.g., name; can be empty):
    Alice <alice@example.com>
    secret key file created:
    /home/frank/.config/codechain/secrets/KDKOGoY8ErjOnbDQb4k8SZFMvWdAIb-x6FGKKCRby70
    public key with signature and optional comment:
    KDKOGoY8ErjOnbDQb4k8SZFMvWdAIb-x6FGKKCRby70 JNBIdjLOu20He3c-Dn7sjpspO8bmKFxTlOItfZkqieb8h218t3g-QooDATGGbrzYzVNbDqb7LCFFnJxEH7hcBA 'Alice <alice@example.com>'

Let's start using Codechain for our example project:

    $ codechain start -s ~/.config/codechain/secrets/KDKOGoY8ErjOnbDQb4k8SZFMvWdAIb-x6FGKKCRby70
    passphrase: 
    e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 2018-05-19T00:07:02Z cstart KDKOGoY8ErjOnbDQb4k8SZFMvWdAIb-x6FGKKCRby70 sVnVenzHyCOV6nLUkCKg6ARllkYsTV-n 0UmUcDFZ2j3WWnqzEdxX-wzofWlhF3O0Rm1tT6qMUwLu8a1R5MwbK5zDongYZKccpA37Vp6Sp3m0xSreGskzCg Alice <alice@example.com>

Let's add Bob (who already has a key) as reviewer:

    $ codechain addkey 91HOu2fvkjHd5S0LtAWTl6dYBk5cqB-NWiJqc0c_7Gc Xsr_L-1_5_B56vocve8s3Pb3vJoc-jpa2-tzIQhEjuoytYfcAiONu3er6RnVNMcsPuZFeqWCQKBwka-F-c13Ag 'Bob <bob@example.com>'
    40c7e5ca4be98e9cae6931afa4ac09e11ecb1ce20fa18d0faaabfac7e8fad071 2018-05-19T00:09:44Z addkey 1 91HOu2fvkjHd5S0LtAWTl6dYBk5cqB-NWiJqc0c_7Gc Xsr_L-1_5_B56vocve8s3Pb3vJoc-jpa2-tzIQhEjuoytYfcAiONu3er6RnVNMcsPuZFeqWCQKBwka-F-c13Ag Bob <bob@example.com>

Increase number of necessary signers to two:

    $ codechain sigctl -m 2
    34cd10effd93e67ba96fefb29ea751d013459a6de11cc117cf1deacd77d6b7be 2018-05-19T00:10:25Z sigctl 2

Publish first release:

    $ codechain publish
    opening keyfile: /home/frank/.config/codechain/secrets/KDKOGoY8ErjOnbDQb4k8SZFMvWdAIb-x6FGKKCRby70
    passphrase: 
    diff --git a/.codechain/tree/b/README.md b/.codechain/tree/b/README.md
    new file mode 100644
    index 0000000..841852c
    --- /dev/null
    +++ b/.codechain/tree/b/README.md
    @@ -0,0 +1 @@
    +## Example project for Codechain walkthrough
    diff --git a/.codechain/tree/b/hello.go b/.codechain/tree/b/hello.go
    new file mode 100644
    index 0000000..c40eee0
    --- /dev/null
    +++ b/.codechain/tree/b/hello.go
    @@ -0,0 +1,9 @@
    +package main
    +
    +import (
    +       "fmt"
    +)
    +
    +func main() {
    +       fmt.Println("hello world!")
    +}
    publish patch? [y/n]: y
    comment describing code change (can be empty):
    first release
    92d2fc6687b0d36d045adaf34a1615e513ef0e2dc60384cfe19863e9753567f8 2018-05-19T00:11:44Z source d844cbe6f6c2c29e97742b272096407e4d92e6ac7f167216b321c7aa55629716 KDKOGoY8ErjOnbDQb4k8SZFMvWdAIb-x6FGKKCRby70 r5aZCYGwWCFppaMDV7XSOHoyCl3qbUKGiSuYzjsTl4C0W9n0tCa0MXDy_fOwspV9f4_o0kMcb6XZS706ml3FAQ first release

Review changes:

    $ codechain review
    opening keyfile: /home/frank/.config/codechain/secrets/KDKOGoY8ErjOnbDQb4k8SZFMvWdAIb-x6FGKKCRby70
    passphrase: 
    signer/sigctl changes:
    0 addkey 1 91HOu2fvkjHd5S0LtAWTl6dYBk5cqB-NWiJqc0c_7Gc Bob <bob@example.com>
    0 sigctl 2
    confirm signer/sigctl changes? [y/n]: y
    patch 1/1
    first release
    developer: KDKOGoY8ErjOnbDQb4k8SZFMvWdAIb-x6FGKKCRby70
    Alice <alice@example.com>
    review patch (no aborts)? [y/n]: y
    diff --git a/.codechain/tree/b/README.md b/.codechain/tree/b/README.md
    new file mode 100644
    index 0000000..841852c
    --- /dev/null
    +++ b/.codechain/tree/b/README.md
    @@ -0,0 +1 @@
    +## Example project for Codechain walkthrough
    diff --git a/.codechain/tree/b/hello.go b/.codechain/tree/b/hello.go
    new file mode 100644
    index 0000000..c40eee0
    --- /dev/null
    +++ b/.codechain/tree/b/hello.go
    @@ -0,0 +1,9 @@
    +package main
    +
    +import (
    +       "fmt"
    +)
    +
    +func main() {
    +       fmt.Println("hello world!")
    +}
    sign patch? [y/n]: y
    d258ce20943beeed2d483096702a1449447f112dec7d907d50c285c649c17a24 2018-05-19T00:12:48Z signtr d258ce20943beeed2d483096702a1449447f112dec7d907d50c285c649c17a24 KDKOGoY8ErjOnbDQb4k8SZFMvWdAIb-x6FGKKCRby70 HKlLKnYSCVzc4b-erETK50EN5gKRKZQsT16grv7eFBklFqXBFoSXSmcY99HLWhAP9BJcA6c3Px1trNBns3KkDA

See current status of project:

    $ codechain status
    no signed releases yet

    signers (2-of-2 required):
    1 91HOu2fvkjHd5S0LtAWTl6dYBk5cqB-NWiJqc0c_7Gc Bob <bob@example.com>
    1 KDKOGoY8ErjOnbDQb4k8SZFMvWdAIb-x6FGKKCRby70 Alice <alice@example.com>

    unsigned entries:
    1 source d844cbe6f6c2c29e97742b272096407e4d92e6ac7f167216b321c7aa55629716 first release

    head:
    2e34e23ee293e8c0ed174639d325eb3e30f5337d5c5846380367724e93cb619e

    tree matches d844cbe6f6c2c29e97742b272096407e4d92e6ac7f167216b321c7aa55629716

Because Bob was added as the second reviewer first, we still need his
signature for the first release. Let's build a distribution for him:

    $ codechain createdist -f /tmp/dist.tar.gz

Now as Bob, apply the distribution in an empty `~/helloproject`
directory:

    $ cd ~/helloproject
    $ codechain apply -f /tmp/dist.tar.gz
    $ find . -type f
    ./.codechain/hashchain
    ./.codechain/patches/e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855

Now Bob reviews the changes and creates a detached signature:

    $ codechain review -d
    opening keyfile: /home/frank/bob.bin
    passphrase: 
    patch 1/1
    first release
    developer: KDKOGoY8ErjOnbDQb4k8SZFMvWdAIb-x6FGKKCRby70
    Alice <alice@example.com>
    review patch (no aborts)? [y/n]: y
    diff --git a/.codechain/tree/b/README.md b/.codechain/tree/b/README.md
    new file mode 100644
    index 0000000..841852c
    --- /dev/null
    +++ b/.codechain/tree/b/README.md
    @@ -0,0 +1 @@
    +## Example project for Codechain walkthrough
    diff --git a/.codechain/tree/b/hello.go b/.codechain/tree/b/hello.go
    new file mode 100644
    index 0000000..c40eee0
    --- /dev/null
    +++ b/.codechain/tree/b/hello.go
    @@ -0,0 +1,9 @@
    +package main
    +
    +import (
    +       "fmt"
    +)
    +
    +func main() {
    +       fmt.Println("hello world!")
    +}
    sign patch? [y/n]: y
    2e34e23ee293e8c0ed174639d325eb3e30f5337d5c5846380367724e93cb619e 91HOu2fvkjHd5S0LtAWTl6dYBk5cqB-NWiJqc0c_7Gc xffZultos-MCbI4cNzAzAoccuDSnpL2nq_BsQanIruYM3RXoD9kdC6WiPEUkxrphKdG742IgBWlB3LwY0i1ZCw

Now Alice can add the detached signature:

    $ codechain review -a 2e34e23ee293e8c0ed174639d325eb3e30f5337d5c5846380367724e93cb619e 91HOu2fvkjHd5S0LtAWTl6dYBk5cqB-NWiJqc0c_7Gc xffZultos-MCbI4cNzAzAoccuDSnpL2nq_BsQanIruYM3RXoD9kdC6WiPEUkxrphKdG742IgBWlB3LwY0i1ZCw
    2e34e23ee293e8c0ed174639d325eb3e30f5337d5c5846380367724e93cb619e 2018-05-19T00:34:51Z signtr 2e34e23ee293e8c0ed174639d325eb3e30f5337d5c5846380367724e93cb619e 91HOu2fvkjHd5S0LtAWTl6dYBk5cqB-NWiJqc0c_7Gc xffZultos-MCbI4cNzAzAoccuDSnpL2nq_BsQanIruYM3RXoD9kdC6WiPEUkxrphKdG742IgBWlB3LwY0i1ZCw

Which gives us our first signed release:

    $ codechain status
    signed releases:
    d844cbe6f6c2c29e97742b272096407e4d92e6ac7f167216b321c7aa55629716 first release

    signers (2-of-2 required):
    1 91HOu2fvkjHd5S0LtAWTl6dYBk5cqB-NWiJqc0c_7Gc Bob <bob@example.com>
    1 KDKOGoY8ErjOnbDQb4k8SZFMvWdAIb-x6FGKKCRby70 Alice <alice@example.com>

    no unsigned entries

    head:
    9f97737b292f66e52c06027871be328006f125a9d86fbe1fc4f03ff98303e36f

    tree matches d844cbe6f6c2c29e97742b272096407e4d92e6ac7f167216b321c7aa55629716

Which we can publish now:

    $ codechain createdist -f /tmp/helloproject.tar.gz

Users can apply it a directory `~/helloproject` and verify the hash
chain contains the head with:

    $ cd ~/helloproject
    $ codechain apply -f /tmp/helloproject.tar.gz -head 9f97737b292f66e52c06027871be328006f125a9d86fbe1fc4f03ff98303e36f

The tree hash now matches the first signed release:

    $ codechain treehash
    d844cbe6f6c2c29e97742b272096407e4d92e6ac7f167216b321c7aa55629716

Show the complete hash chain:

    $ codechain status -p
    e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 2018-05-19T00:07:02Z cstart KDKOGoY8ErjOnbDQb4k8SZFMvWdAIb-x6FGKKCRby70 sVnVenzHyCOV6nLUkCKg6ARllkYsTV-n 0UmUcDFZ2j3WWnqzEdxX-wzofWlhF3O0Rm1tT6qMUwLu8a1R5MwbK5zDongYZKccpA37Vp6Sp3m0xSreGskzCg Alice <alice@example.com>
    40c7e5ca4be98e9cae6931afa4ac09e11ecb1ce20fa18d0faaabfac7e8fad071 2018-05-19T00:09:44Z addkey 1 91HOu2fvkjHd5S0LtAWTl6dYBk5cqB-NWiJqc0c_7Gc Xsr_L-1_5_B56vocve8s3Pb3vJoc-jpa2-tzIQhEjuoytYfcAiONu3er6RnVNMcsPuZFeqWCQKBwka-F-c13Ag Bob <bob@example.com>
    34cd10effd93e67ba96fefb29ea751d013459a6de11cc117cf1deacd77d6b7be 2018-05-19T00:10:25Z sigctl 2
    92d2fc6687b0d36d045adaf34a1615e513ef0e2dc60384cfe19863e9753567f8 2018-05-19T00:11:44Z source d844cbe6f6c2c29e97742b272096407e4d92e6ac7f167216b321c7aa55629716 KDKOGoY8ErjOnbDQb4k8SZFMvWdAIb-x6FGKKCRby70 r5aZCYGwWCFppaMDV7XSOHoyCl3qbUKGiSuYzjsTl4C0W9n0tCa0MXDy_fOwspV9f4_o0kMcb6XZS706ml3FAQ first release
    d258ce20943beeed2d483096702a1449447f112dec7d907d50c285c649c17a24 2018-05-19T00:12:48Z signtr d258ce20943beeed2d483096702a1449447f112dec7d907d50c285c649c17a24 KDKOGoY8ErjOnbDQb4k8SZFMvWdAIb-x6FGKKCRby70 HKlLKnYSCVzc4b-erETK50EN5gKRKZQsT16grv7eFBklFqXBFoSXSmcY99HLWhAP9BJcA6c3Px1trNBns3KkDA
    2e34e23ee293e8c0ed174639d325eb3e30f5337d5c5846380367724e93cb619e 2018-05-19T00:34:51Z signtr 2e34e23ee293e8c0ed174639d325eb3e30f5337d5c5846380367724e93cb619e 91HOu2fvkjHd5S0LtAWTl6dYBk5cqB-NWiJqc0c_7Gc xffZultos-MCbI4cNzAzAoccuDSnpL2nq_BsQanIruYM3RXoD9kdC6WiPEUkxrphKdG742IgBWlB3LwY0i1ZCw
