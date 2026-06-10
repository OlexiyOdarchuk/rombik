def toxic_routing_test(matrix):
    result = 0
    for row in matrix:
        if not row:
            continue

        for val in row:
            if val < 0:
                break
            if val == 999:
                return -1

            if val % 2 == 0:
                result += val
            else:
                while val > 0:
                    val -= 1
                    if val == 5:
                        continue
                    if result > 100:
                        break
        else:
            result += 1000
            continue

    return result
