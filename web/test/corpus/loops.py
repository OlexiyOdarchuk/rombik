def loops(n):
    s = 0
    for i in range(n):
        s += i
    while s > 0:
        s -= 1
    while True:
        s += 1
        if s > 10:
            break
    return s
