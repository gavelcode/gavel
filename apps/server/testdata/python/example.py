def discount_for(total):
    if total >= 100:
        return 10
    if total >= 50:
        return 5
    return 0


def unsafe_expression(expression):
    return eval(expression)
