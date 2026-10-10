#if defined(_WIN32)
#define YS_EXPORT extern "C" __declspec(dllexport)
#else
#define YS_EXPORT extern "C" __attribute__((visibility("default")))
#endif

using ys_i32 = signed int;
using ys_u64 = unsigned long long;

struct ys_pair {
    ys_i32 count;
    double weight;
};

using ys_callback = ys_i32 (*)(ys_i32);

YS_EXPORT ys_i32 ys_lstd_add(ys_i32 left, ys_i32 right) {
    return left + right;
}

YS_EXPORT double ys_lstd_weighted(double value, ys_i32 multiplier) {
    return value * multiplier;
}

YS_EXPORT ys_u64 ys_lstd_strlen(const char* text) {
    ys_u64 length = 0;
    while (text && text[length] != '\0') {
        ++length;
    }
    return length;
}

YS_EXPORT const char* ys_lstd_greeting() {
    return "hello from C++";
}

YS_EXPORT ys_pair ys_lstd_pair_transform(ys_pair input) {
    return {input.count + 5, input.weight * 2.0};
}

YS_EXPORT ys_i32 ys_lstd_apply(ys_callback callback, ys_i32 input) {
    return callback ? callback(input) : -1;
}

YS_EXPORT ys_i32 ys_lstd_upper_first(char* text) {
    if (!text || text[0] == '\0') {
        return -1;
    }
    if (text[0] >= 'a' && text[0] <= 'z') {
        text[0] = static_cast<char>(text[0] - ('a' - 'A'));
    }
    return static_cast<unsigned char>(text[0]);
}

YS_EXPORT ys_i32 ys_lstd_ordinal_value() {
    return 1701;
}
