int sum(int n) {
    int s = 0;
    for (int i = 0; i < n; i++) {
        if (i % 2 == 0) {
            s += i;
        }
    }
    return s;
}

int main() {
    int n;
    cin >> n;
    cout << sum(n) << endl;
    return 0;
}
