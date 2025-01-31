package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// model, следующая строка
// Определяет структуру данных для хранения состояния приложения.
// Как работает функция: Эта структура содержит все необходимые поля для управления состоянием калькулятора, включая ввод пользователя, текущее выражение и результаты.
type model struct {
	expression   string
	result       string
	intermediate string
	input        textinput.Model
}

// initialModel, следующая строка
// Инициализирует начальное состояние приложения.
// Как работает функция: Создает и возвращает экземпляр модели с начальными значениями.
func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Введите выражение"
	ti.Focus()
	ti.CharLimit = 50
	ti.Width = 30

	return model{
		expression:   "",
		result:       "",
		intermediate: "",
		input:        ti,
	}
}

// Init, следующая строка
// Инициализирует приложение.
// Как работает функция: Возвращает начальную команду для запуска приложения.
func (m model) Init() tea.Cmd {
	return textinput.Blink
}

// Update, следующая строка
// Обрабатывает обновления состояния приложения.
// Как работает функция: Принимает сообщения и обновляет состояние модели в зависимости от типа сообщения.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			m.expression = m.input.Value()
			m.result, m.intermediate = evaluateExpression(m.expression)
			m.input.Reset()
			return m, nil
	case tea.KeyCtrlC, tea.KeyEsc:
		return m, tea.Quit
		default:

		}
	}

	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// View, следующая строка
// Отображает текущее состояние приложения.
// Как работает функция: Формирует и возвращает строку, которая будет отображаться в терминале.
func (m model) View() string {
	return fmt.Sprintf(
		"%s\n%s\n%s\n%s",
		lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render("Результат: "+m.result),
		lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Render("Промежуточный: "+m.intermediate),
		m.input.View(),
		lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("Используйте Enter для вычисления, Ctrl+C для выхода."),
	)
}

// evaluateExpression, следующая строка
// Вычисляет значение математического выражения.
// Как работает функция: Разбирает строку выражения, вычисляет его значение и возвращает результат и промежуточный результат.
func evaluateExpression(expr string) (string, string) {
	// Удаляем все пробелы из выражения
	expr = strings.ReplaceAll(expr, " ", "")

	// Проверяем, что выражение состоит только из допустимых символов
	for _, char := range expr {
		if !unicode.IsDigit(char) && !strings.ContainsRune("+-*/().^√", char) {
			return "Ошибка", "Некорректное выражение"
		}
	}

	// Вычисляем выражение
	result, err := calculate(expr)
	if err != nil {
		return "Ошибка", err.Error()
	}

	// Промежуточный результат (например, вычисление без учета скобок)
	intermediate, _ := calculate(strings.ReplaceAll(expr, "(", ""))
	return fmt.Sprintf("%.2f", result), fmt.Sprintf("%.2f", intermediate)
}

// calculate, следующая строка
// Вычисляет значение математического выражения.
// Как работает функция: Использует алгоритм сортировочной станции (Shunting Yard) для преобразования выражения в обратную польскую запись (ОПЗ) и вычисляет его.
func calculate(expr string) (float64, error) {
	// Преобразуем выражение в ОПЗ
	rpn, err := shuntingYard(expr)
	if err != nil {
		return 0, err
	}

	// Вычисляем значение ОПЗ
	result, err := evalRPN(rpn)
	if err != nil {
		return 0, err
	}

	return result, nil
}

// shuntingYard, следующая строка
// Преобразует выражение в обратную польскую запись (ОПЗ).
// Как работает функция: Использует алгоритм сортировочной станции для преобразования инфиксного выражения в ОПЗ.
func shuntingYard(expr string) ([]string, error) {
	var output []string
	var operators []string

	for i := 0; i < len(expr); i++ {
		char := rune(expr[i])

		if unicode.IsDigit(char) || char == '.' {
			num := ""
			for i < len(expr) && (unicode.IsDigit(rune(expr[i])) || expr[i] == '.') {
				num += string(expr[i])
				i++
			}
			i--
			output = append(output, num)
		} else if strings.ContainsRune("+-*/^√", char) {
			for len(operators) > 0 && precedence(operators[len(operators)-1]) >= precedence(string(char)) {
				output = append(output, operators[len(operators)-1])
				operators = operators[:len(operators)-1]
			}
			operators = append(operators, string(char))
		} else if char == '(' {
			operators = append(operators, string(char))
		} else if char == ')' {
			for len(operators) > 0 && operators[len(operators)-1] != "(" {
				output = append(output, operators[len(operators)-1])
				operators = operators[:len(operators)-1]
			}
			if len(operators) == 0 {
				return nil, fmt.Errorf("непарные скобки")
			}
			operators = operators[:len(operators)-1]
		}
	}

	for len(operators) > 0 {
		if operators[len(operators)-1] == "(" || operators[len(operators)-1] == ")" {
			return nil, fmt.Errorf("непарные скобки")
		}
		output = append(output, operators[len(operators)-1])
		operators = operators[:len(operators)-1]
	}

	return output, nil
}

// precedence, следующая строка
// Определяет приоритет операторов.
// Как работает функция: Возвращает числовое значение приоритета оператора.
func precedence(op string) int {
	switch op {
	case "+", "-":
		return 1
	case "*", "/":
		return 2
	case "^", "√":
		return 3
	default:
		return 0
	}
}

// evalRPN, следующая строка
// Вычисляет значение выражения в обратной польской записи.
// Как работает функция: Использует стек для вычисления значения ОПЗ.
func evalRPN(rpn []string) (float64, error) {
	var stack []float64

	for _, token := range rpn {
		if num, err := strconv.ParseFloat(token, 64); err == nil {
			stack = append(stack, num)
		} else {
			if len(stack) < 2 {
				return 0, fmt.Errorf("недостаточно операндов")
			}
			b := stack[len(stack)-1]
			a := stack[len(stack)-2]
			stack = stack[:len(stack)-2]

			switch token {
			case "+":
				stack = append(stack, a+b)
			case "-":
				stack = append(stack, a-b)
			case "*":
				stack = append(stack, a*b)
			case "/":
				if b == 0 {
					return 0, fmt.Errorf("деление на ноль")
				}
				stack = append(stack, a/b)
			case "^":
				stack = append(stack, pow(a, b))
			case "√":
				if a < 0 {
					return 0, fmt.Errorf("корень из отрицательного числа")
				}
				stack = append(stack, sqrt(a))
			default:
				return 0, fmt.Errorf("неизвестный оператор: %s", token)
			}
		}
	}

	if len(stack) != 1 {
		return 0, fmt.Errorf("ошибка вычисления")
	}

	return stack[0], nil
}

// pow, следующая строка
// Вычисляет степень числа.
// Как работает функция: Возвращает a в степени b.
func pow(a, b float64) float64 {
	result := 1.0
	for i := 0; i < int(b); i++ {
		result *= a
	}
	return result
}

// sqrt, следующая строка
// Вычисляет квадратный корень числа.
// Как работает функция: Возвращает квадратный корень из a.
func sqrt(a float64) float64 {
	return math.Sqrt(a) 
}

// main, следующая строка
// Запускает приложение.
// Как работает функция: Создает и запускает новое приложение Bubble Tea.
func main() {
	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		fmt.Printf("Ошибка: %v", err)
	}
}
