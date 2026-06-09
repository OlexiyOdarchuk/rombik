with open("pkg/layout/layout.go", "r") as f:
    lines = f.readlines()
for i, line in enumerate(lines):
    if "func (b *build) routeEnds" in line:
        start = i
        break
for i in range(start, len(lines)):
    if lines[i].startswith("}"):
        end = i
        break
print("".join(lines[start:end+1]))
