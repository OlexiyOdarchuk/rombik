def grade(score):
    name = input("Ваше ім'я: ")
    print("Привіт,", name)
    total = score + 5
    if total >= 90:
        print("Відмінно")
    else:
        if total >= 60:
            print("Задовільно")
        else:
            print("Незадовільно")
    print("Готово")
