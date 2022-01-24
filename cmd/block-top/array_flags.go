package main

type ArrayFlag []string

func (i *ArrayFlag) String() string {
	return "my string representation"
}

func (i *ArrayFlag) Set(value string) error {
	*i = append(*i, value)
	return nil
}
