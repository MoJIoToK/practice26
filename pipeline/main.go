package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"
)

// RingIntBuffer - структура кольцевого буфера целых чисел. Реализована с помощью слайса целых чисел.
// Полями структуры являются: array - слайс целых чисел, pos - позиция текущего элемента, size - размер слайса,
// m - мьютекс для синхронизации операций над буфером с помощью нескольких горутин.
type RingIntBuffer struct {
	array []int
	pos   int
	size  int
	m     sync.Mutex
}

// NewRingIntBuffer - конструктор. Принимает размер буфера, создает объект структуры и возвращает указатель на
// созданный объект.
func NewRingIntBuffer(size int) *RingIntBuffer {
	return &RingIntBuffer{make([]int, size), -1, size, sync.Mutex{}}
}

// Push - позволяет записать элемент в буфер. На вход подаётся целочисленный элемент, проверяется заполнен ли буфер.
// Если буфер заполнен, то самое старое значение стирается.
func (r *RingIntBuffer) Push(elem int) {
	//Закрытие структуры для записи и чтения
	r.m.Lock()
	defer r.m.Unlock()
	//Проверка заполнился ли буфер или нет
	if r.pos == r.size-1 {
		for i := 1; i <= r.size-1; i++ {
			//Сдвиг всех элементов на одну позицию в начало. Т.е. затирание самого старого значения.
			r.array[i-1] = r.array[i]
		}
		r.array[r.pos] = elem
	} else {
		//Увеличение текущей позиции буфера.
		r.pos++
		r.array[r.pos] = elem
	}

}

// Get - позволяет получить элемент из буфера. Метод возвращает слайс
func (r *RingIntBuffer) Get() []int {
	if r.pos <= 0 {
		return nil
	}
	r.m.Lock()
	defer r.m.Unlock()
	var output []int = r.array[:r.pos+1]

	r.pos = -1
	return output
}

// read - метод позволяющий читать данные с консоли и записывать их в канал.
func read(input chan<- int) {
	for {
		var u int
		_, err := fmt.Scanf("%d\n", &u)
		if err != nil {
			fmt.Println("This is not the number!")
		}
		input <- u
	}
}

// removeNegatives - метод реализует стадию фильтрации данных в канале чисел меньше нуля.
func removeNegatives(currentChannel <-chan int, nextChan chan<- int) {
	for number := range currentChannel {
		if number >= 0 {
			nextChan <- number
		}
	}
}

// removeDivThree - метод реализует стадию фильтрации данных канала не кратных 3. Исключая 0.
func removeDivThree(currentChannel <-chan int, nextChan chan<- int) {
	for number := range currentChannel {
		if number%3 == 0 {
			nextChan <- number
		}
	}
}

// writeToBuffer - метод записывающий данные канала в кольцевой буфер. По сути обертка метода Push() для каждого
// элемента в канале.
func writeToBuffer(currentChannel <-chan int, r *RingIntBuffer) {
	for number := range currentChannel {
		r.Push(number)
	}
}

// writeToConsole - метод считывающие данные из кольцевого буфера и выводящий в консоль.
func writeToConsole(r *RingIntBuffer, ticker *time.Ticker) {
	for range ticker.C {
		buffer := r.Get()
		if len(buffer) > 0 {
			fmt.Println("The buffer is", buffer)
		}
	}
}

func main() {

	input := make(chan int)
	go read(input)

	//Запуск стадии фильтрации отрицательных чисел.
	negFilterChannel := make(chan int)
	go removeNegatives(input, negFilterChannel)

	//Запуск стадии фильтрации чисел кратных 3.
	divThreeChannel := make(chan int)
	go removeDivThree(negFilterChannel, divThreeChannel)

	size := 20
	r := NewRingIntBuffer(size)
	go writeToBuffer(divThreeChannel, r)

	delay := 5
	ticker := time.NewTicker(time.Second * time.Duration(delay))
	go writeToConsole(r, ticker)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	select {
	case sig := <-c:
		fmt.Println("Got %s signal. Aborting ...", sig)
		os.Exit(0)

	}

}
