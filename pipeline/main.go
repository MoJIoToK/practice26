package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
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

// Get - позволяет получить элемент из буфера. Метод возвращает слайс.
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
	log.Printf("Start read stage!")
	scanner := bufio.NewScanner(os.Stdin)
	var data string
	for scanner.Scan() {
		data = scanner.Text()
		log.Printf("Data input %v", data)
		if strings.EqualFold(data, "exit") {
			fmt.Println("Program exits!")
			log.Printf("Сlosing the program by the user")
			return
		}
		number, err := strconv.Atoi(data)
		log.Printf("Text %v transformed to digit", data)
		if err != nil {
			log.Printf("Error! Entered not a number!")
			fmt.Println("This is not the number!")
		}
		input <- number
		log.Printf("End read stage!")
	}
}

// removeNegatives - метод реализует стадию фильтрации данных в канале чисел меньше нуля.
func removeNegatives(currentChannel <-chan int, nextChan chan<- int) {
	log.Printf("Start negative filter stage!")
	for number := range currentChannel {
		if number >= 0 {
			log.Printf("Number more than or equals 0!")
			nextChan <- number
		}
	}
	log.Printf("End negative filter stage!")
}

// removeDivThree - метод реализует стадию фильтрации данных канала не кратных 3. Исключая 0.
func removeDivThree(currentChannel <-chan int, nextChan chan<- int) {
	log.Printf("Start div 3 filter stage!")
	for number := range currentChannel {
		if number%3 == 0 {
			log.Printf("Number div 3!")
			nextChan <- number
		}
	}
	log.Printf("End div 3 filter stage!")
}

// writeToBuffer - метод записывающий данные канала в кольцевой буфер. По сути обертка метода Push() для каждого
// элемента в канале.
func writeToBuffer(currentChannel <-chan int, r *RingIntBuffer) {
	log.Printf("Start write number in ringBuffer stage!")
	for number := range currentChannel {
		r.Push(number)
	}
	log.Printf("End write number in ringBuffer stage!")
}

// writeToConsole - метод считывающие данные из кольцевого буфера и выводящий в консоль.
func writeToConsole(r *RingIntBuffer, ticker *time.Ticker) {
	log.Printf("Start write number in console stage!")
	for range ticker.C {
		buffer := r.Get()
		if len(buffer) > 0 {
			log.Printf("Output data to the console!")
			fmt.Println("The buffer is", buffer)
		}
	}
	log.Printf("End write number in console stage!")
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
		fmt.Printf("Got %s signal. Aborting ...", sig)
		os.Exit(0)

	}

}
