# block top

Continuously watch the block production from one or many (or all) leaders on solana.

Intended for quick and dirty observation of the block production statistics of a particular leader (or all).

Built on code from https://github.com/certusone/tpuproxy/blob/main/cmd/txcount/. 


## Usage

**Watch all leaders:**

`go run ./cmd/block-top` 

**Watch single key**

`go run ./cmd/block-top -leader DDnAqxJVFo2GVTujibHt5cjevHMSE9bo8HJaydHoshdp` 

**Watch multiple keys**

`go run ./cmd/block-top -leader DDnAqxJVFo2GVTujibHt5cjevHMSE9bo8HJaydHoshdp -leader Certusm1sa411sMpV9FPqU5dXAYhmmhygvxJ23S6hJ24`  


Pipe stderr to `/dev/null` to hide debug output.



