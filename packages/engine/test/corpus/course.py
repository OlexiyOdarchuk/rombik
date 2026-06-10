import math
import random
from typing import Callable


def input_variables[T](
    convert_func: Callable[[str], T],
    message: str = "DEFAULT",
    digit_count: int = 0,
    min_val: int | float | None = None,
    max_val: int | float | None = None,
    amount: int = 1,
    rows: int = 1,
    cols: int = 0,
    validator: Callable[[T], bool] | None = None,
) -> list[list[T]]:
    """
    Універсальний інструмент для зчитування та валідації даних з консолі.

    Функція забезпечує типізоване введення, підтримує створення матриць,
    перевірку діапазонів значень, довжини чисел та дозволяє передавати
    власні функції-валідатори для складної логіки.

    Args:
        convert_func: Функція для перетворення введеного рядка у тип T (напр. int, float).
        message: Рядок підказки для користувача. "DEFAULT" генерує автоматичне повідомлення.
        digit_count: Вимагає, щоб кожне число складалося саме з цієї кількості цифр.
        min_val: Мінімально допустиме значення для числових типів.
        max_val: Максимально допустиме значення для числових типів.
        amount: Кількість елементів, які очікуються в одному рядку.
        rows: Кількість рядків для зчитування.
        cols: Кількість колонок у матриці (якщо 0, використовується значення amount).
        validator: Опціональна функція, що приймає значення T і повертає True, якщо воно валідне.

    Returns:
        Двовимірний список (list[list[T]]).

    Raises:
        ValueError: Якщо параметр amount менший за 1.

    Example:
        >>> # Зчитати 3 парних числа від 10 до 99
        >>> nums = input_variables(int, amount=3, min_val=10, max_val=99, validator=lambda x: x % 2 == 0)
    """

    if amount < 1:
        raise ValueError("the number of elements (amount) must be >= 1")

    actual_cols = cols if cols > 0 else amount

    def process_line(line: str, expected_count: int) -> list[T]:
        parts = line.split()
        if len(parts) < expected_count:
            raise ValueError(
                f"not enough data; expected {expected_count}, received {len(parts)}"
            )

        parts = parts[:expected_count]

        try:
            parsed = [convert_func(p) for p in parts]
        except Exception as e:
            raise ValueError(e)

        for i, val in enumerate(parsed):
            if digit_count > 0:
                clean_str = str(parts[i]).lstrip("-").replace(".", "")
                if len(clean_str) != digit_count:
                    raise ValueError(
                        f"the number '{parts[i]}' must consist of {digit_count} digits"
                    )

            if isinstance(val, (int, float)):
                if min_val is not None and val < min_val:
                    raise ValueError(
                        f"the value of {val} is less than the minimum ({min_val})"
                    )
                if max_val is not None and val > max_val:
                    raise ValueError(
                        f"the value of {val} is greater than the maximum ({max_val})"
                    )

            if validator and not validator(val):
                raise ValueError(f"the value {val} did not pass validation")

        return parsed

    result: list[list[T]] = []
    needed_rows = rows

    for r in range(needed_rows):
        while True:
            if message == "DEFAULT":
                prompt = (
                    f"Рядок {r + 1}/{rows} (очікується {actual_cols}): "
                    if rows > 1
                    else "Введіть значення: "
                )
            else:
                prompt = message

            try:
                row = process_line(input(prompt), actual_cols)
                result.append(row)
                break
            except ValueError as e:
                print(f"⚠️ {e}. Спробуйте ще раз.")

    return result


class _RowProxy[T]:
    """Допоміжний клас для обробки другого індексу matrix[i][j]"""

    def __init__(self, row_ref: list[T | None], cols_count: int) -> None:
        self.row: list[T | None] = row_ref
        self.cols: int = cols_count

    def __getitem__(self, j: int) -> T | None:
        if not (1 <= j <= self.cols):
            raise IndexError(f"column index {j} out of range [1..{self.cols}]")
        return self.row[j - 1]

    def __setitem__(self, j: int, value: T) -> None:
        if not (1 <= j <= self.cols):
            raise IndexError(f"column index {j} out of range [1..{self.cols}]")
        self.row[j - 1] = value


class Matrix1Based[T]:
    """
    Матриця з індексацією від 1 для завдання 5.
    """

    def __init__(self, rows: int, cols: int, default_value: T | None = None) -> None:
        self.rows: int = rows
        self.cols: int = cols
        self._data: list[list[T | None]] = [
            [default_value for _ in range(cols)] for _ in range(rows)
        ]

    def __getitem__(self, i: int) -> _RowProxy[T]:
        if not (1 <= i <= self.rows):
            raise IndexError(f"row index {i} out of range [1..{self.rows}]")

        return _RowProxy(self._data[i - 1], self.cols)

    def __repr__(self) -> str:
        """Форматований вивід таблиці з українськими назвами та вирівнюванням."""
        if not self._data:
            return "Матриця порожня"

        col_width = 12
        row_label_width = 10
        res: list[str] = []

        header = " " * (row_label_width + 2)
        for j in range(1, self.cols + 1):
            header += f"{f'Стовп. {j}':>{col_width}}"
        res.append(header)

        separator = " " * row_label_width + " ┌" + "─" * (self.cols * col_width)
        res.append(separator)

        for i, row in enumerate(self._data, 1):
            row_label = f"Рядок {i:<2}"
            line = f"{row_label:>{row_label_width}} │"

            for item in row:
                if item is not None:
                    line += f"{item:>{col_width}.3f}"
                else:
                    line += f"{'---':>{col_width}}"
            res.append(line)

        return "\n".join(res)

    def get(self) -> list[list[T | None]]:
        """Повертає raw-дані для обробки."""
        return self._data


def main():
    print("--- Курсова робота ---")
    print("--- Варіант 21 ---")

    tasks = [
        ("Завдання 1", first_task, float, 3),
        ("Завдання 2", second_task, float, 1),
        ("Завдання 3", third_task, float, 1),
        ("Завдання 4", fourth_task, int, 1),
    ]

    for title, func, data_type, args_count in tasks:
        print(f"\n--- {title} ---")

        data: list[list] = input_variables(data_type, amount=args_count)

        result = func(*data[0])

        print(f"Результат виконання {title.lower()}: {result}")

    print("\n--- Завдання 5 ---")
    fifth_task()


def first_task(x: float, y: float, z: float) -> float:
    """Перше завдання

    Args:
        x (float): змінна X, число з клавіатури
        y (float): змінна Y, число з клавіатури
        z (float): змінна Z, число з клавіатури

    Returns:
        float: результат виконання обчислення, описаного в 1 завданні курсової роботи для варіанту 21
    """
    if x <= 0:
        print("Помилка: x має бути більше 0 для обчислення ln(x)")
        return 0

    term1: float = y - z
    numerator: float = y - (z ** (y - z))
    denominator: float = 1 + x**2
    ln_x: float = math.log(x)

    f: float = term1 * (numerator / denominator) + ln_x
    return f


def second_task(x: float) -> float:
    """Друге завдання

    Args:
        x (float): число з клавіатури

    Returns:
        float: результат виконання обчислення, описаного в 2 завданні курсової роботи для варіанту 21
    """
    a: float = 3.1
    if x > 1:
        return math.sqrt(a + math.log10(x))
    if abs(x) < 1:
        return math.asin(x)
    else:
        return x**a


def third_task(x: float) -> float:
    """Третє завдання

    Args:
        x (float): число з клавіатури

    Raises:
        RuntimeError: Якщо алгоритми повертають не однаковий результат

    Returns:
        float: результат виконання обчислення, описаного в 3 завданні курсової роботи для варіанту 21
    """
    strategies: list[tuple[str, Callable[[float], float]]] = [
        ("Виконання циклу з лічильником", for_cycle_third_task),
        ("Виконання циклу з передумовою", while_cycle_third_task),
        ("Виконання циклу з післяумовою", while_true_cycle_third_task),
    ]
    result: list[float] = []
    for name, func in strategies:
        localResult = func(x)
        result.append(localResult)
        print(f"{name}: {localResult}")

    for i in range(len(result)):
        if not math.isclose(result[i], result[0]):
            raise RuntimeError("Помилка. Результат відрізняється!")

    return result[0]


def for_cycle_third_task(x: float) -> float:
    """Цикл з лічильником для третього завдання

    Args:
        x (float): число з клавіатури

    Returns:
        float: результат виконання обчислення, описаного в 3 завданні курсової роботи для варіанту 21
    """
    sum: float = 0.0
    k: int = 1
    for k in range(1, 7):
        sum += math.sin(0.17 * (x**k)) / ((2 * k) + x)
    return sum


def while_cycle_third_task(x: float) -> float:
    """Цикл з передумовою для третього завдання

    Args:
        x (float): число з клавіатури

    Returns:
        float: результат виконання обчислення, описаного в 3 завданні курсової роботи для варіанту 21
    """
    sum: float = 0.0
    k: int = 1
    while k < 7:
        sum += math.sin(0.17 * (x**k)) / ((2 * k) + x)
        k += 1
    return sum


def while_true_cycle_third_task(x: float) -> float:
    """Цикл з післяумовою для третього завдання

    Args:
        x (float): число з клавіатури

    Returns:
        float: результат виконання обчислення, описаного в 3 завданні курсової роботи для варіанту 21
    """
    sum: float = 0.0
    k: int = 1
    while True:
        sum += math.sin(0.17 * (x**k)) / ((2 * k) + x)
        k += 1
        if k >= 7:
            break
    return sum


def fourth_task(n: int) -> int:
    """Четверте завдання

    Args:
        n (int): число з клавіатури, що показує кількість елементів в списку

    Returns:
        int: сума елементів списку, які є непарними числами
    """
    return sum([i for i in [random.randint(0, 50) for _ in range(n)] if i % 2 != 0])


def fifth_task():
    """Запуск п'ятого завдання"""
    height, width = 5, 5
    matrix = create_matrix_fifth_task(height, width)
    vector = create_vector_fifth_task(matrix)
    print("Створена матриця:")
    print_matrix_fifth_task(matrix)
    print("\nСтворений вектор:")
    print_vector_fifth_task(vector)


def create_matrix_fifth_task(height: int = 5, width: int = 5) -> Matrix1Based[float]:
    """
    Створення матриці для п'ятого завдання.

    Args:
        height: висота матриці (кількість рядків).
        width: ширина матриці (кількість стовпців).

    Returns:
        Matrix1Based[float]: об'єкт матриці з обчисленими значеннями.
    """
    matrix = Matrix1Based[float](height, width)

    for i in range(1, height + 1):
        for j in range(1, width + 1):
            value: float = ((3 + i) / (i + j)) * math.sqrt(i**3 + j**2) + 2 ** (i - j)
            matrix[i][j] = value

    return matrix


def create_vector_fifth_task(matrix: Matrix1Based[float]) -> list[float]:
    """
    Створення вектора на основі сум непарних рядків матриці.

    Args:
        matrix: Об'єкт Matrix1Based, з якого беруться дані.

    Returns:
        list[float]: список сум (для рядків 1, 3, 5).
    """
    vector: list[float] = []

    for i in range(1, matrix.rows + 1, 2):
        row_sum = 0.0
        for j in range(1, matrix.cols + 1):
            val = matrix[i][j]
            if val is not None:
                row_sum += val
        vector.append(row_sum)

    return vector


def print_matrix_fifth_task(matrix: Matrix1Based[float]):
    """Вивід матриці

    Args:
        matrix (Matrix1Based[float]): Матриця, яку потрібно буде вивести (виводить з елемента з індексом 0, а не 1)
    """
    print()
    print(matrix)


def print_vector_fifth_task(vector: list[float]):
    """Вивід вектора

    Args:
        vector (list[float]): Вектор, який потрібно вивести
    """
    formatted_elements: list[str] = [f"{item:.3f}" for item in vector]
    print(*formatted_elements)


def vika():
    n = int(input("Введіть n: "))

    a = []
    for i in range(n):
        a.append(random.randint(-10, 10))
    print("Список:", a)

    neg_a = -1
    neg_b = -1
    for i in range(n):
        if a[i] < 0:
            if neg_a == -1:
                neg_a = i
            elif neg_b == -1:
                neg_b = i
                break

    if neg_a == -1 or neg_b == -1:
        print("У списку менше двох від'ємних елементів")
    else:
        s = 0
        for i in range(neg_a + 1, neg_b):
            s = s + a[i]
        print("Сума між першим та другим від'ємними:", s)


if __name__ == "__main__":
    main()
