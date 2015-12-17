package main

type Vector []float64

func (v Vector) DoSome(i, n int, u Vector, c chan int) {
	for ; i < n; i++ {
		v[i] += u.Op(v[i])
	}
	c <- 1
}

const NCPU = 16

func (v Vector) DoAll(u Vector) {
	c = make(chan int, NCPU)

	for i := 0; i < NCPU; i++ {
		go V.DoSome(i*len(v)/NCPU, (i+1)*len(v)/NCPU, u, c)
	}

	for i := 0; i < NCPU; i++ {
		<-c
	}
	//这里表示计算结束
}

func foo() {
	l.Lock()
	defer l.Unlock()
	//..
}
