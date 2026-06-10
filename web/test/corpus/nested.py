def outer(x):
    def helper(y):
        if y < 0:
            raise ValueError("neg")
        return y * 2
    total = 0
    for i in range(x):
        if i % 2 == 0:
            total += helper(i)
        else:
            total -= 1
    return total
