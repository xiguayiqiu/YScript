// ys_cpp.cpp —— C++ 代码
//
// 演示点：
//   - 类 / 模板 / STL 全部可用（C++ 编译器负责）
//   - 但要导给 YScript 调用的符号必须用 extern "C" 包住，
//     否则 C 链接器找不到名字（name mangling）
#include <cstdint>
#include <string>
#include <vector>
#include <algorithm>
#include <numeric>
#include <cstring>
#include <cstdlib>

namespace {

// 一个普通 C++ 类：编译期由编译器自由优化
class Accumulator {
public:
    explicit Accumulator(std::int64_t seed) : total_(seed) {}
    void add(std::int64_t v) { total_ += v; }
    std::int64_t total() const { return total_; }
    std::int64_t count() const { return n_; }
private:
    std::int64_t total_;
    std::int64_t n_ = 0;
};

// 模板函数：实例化后就是一个普通函数
template <typename T>
T square(T v) { return v * v; }

}  // namespace

extern "C" {

// 1) 类：走一遍 add/total
std::int64_t cpp_accumulate(std::int64_t seed, std::int64_t a, std::int64_t b) {
    Accumulator acc(seed);
    acc.add(a);
    acc.add(b);
    return acc.total();
}

// 2) 模板：求平方（整型与浮点各来一次）
std::int64_t cpp_square_i64(std::int64_t v) { return square<std::int64_t>(v); }
double        cpp_square_f64(double v)        { return square<double>(v); }

// 3) STL：std::vector 求和
//    缓冲由 C++ 侧填充（脚本层没有 int64 数组构造能力），再用指针传进 std::vector
void cpp_fill_i64(std::int64_t* buf, std::int64_t n) {
    for (std::int64_t i = 0; i < n; i++) buf[i] = (i + 1) * 10;
}

std::int64_t cpp_vector_sum(const std::int64_t* arr, std::int64_t n) {
    std::vector<std::int64_t> v(arr, arr + n);
    return std::accumulate(v.begin(), v.end(), std::int64_t{0});
}

// 4) STL：排序后取中位数（std::sort）
double cpp_median(const std::int64_t* arr, std::int64_t n) {
    std::vector<std::int64_t> v(arr, arr + n);
    if (v.empty()) return 0.0;
    std::sort(v.begin(), v.end());
    const std::size_t mid = v.size() / 2;
    if (v.size() % 2 == 1) return static_cast<double>(v[mid]);
    return (static_cast<double>(v[mid - 1]) + static_cast<double>(v[mid])) / 2.0;
}

// 4b) 同上但返回整型（乘 2 以保住 .5 的中位数）。
//     之所以再导出一个 int64 版本：当前 FFI 调用层只支持整型/字符串/字节串参数，
//     浮点参数暂未支持，脚本侧只能调这个。
std::int64_t cpp_median_x2(const std::int64_t* arr, std::int64_t n) {
    return static_cast<std::int64_t>(cpp_median(arr, n) * 2.0);
}

// 5) std::string：返回动态分配的字符串，脚本侧读完要 c.free
char* cpp_join(const char* a, const char* b) {
    std::string s = std::string(a) + "-" + std::string(b);
    char* out = static_cast<char*>(malloc(s.size() + 1));
    if (!out) return nullptr;
    std::memcpy(out, s.c_str(), s.size() + 1);
    return out;
}

// 6) 字符串长度（验证 C++ 侧确实处理了 std::string）
std::int64_t cpp_strlen(const std::string& s) {
    return static_cast<std::int64_t>(s.size());
}

}  // extern "C"
