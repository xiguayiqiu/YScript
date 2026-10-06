/* ys_math.c —— 纯 C 基础运算
 *
 * 演示点：
 *   - int64_t 跨语言整型约定（YScript 的 int 就是 64 位有符号）
 *   - 直接调用 libc（sqrt）
 *   - 数组参数：YScript 侧用 bytes 传，内存布局天然兼容
 */
#include <stdint.h>
#include <math.h>

/* C 预处理演示：没定义 YS_FACTOR 时用默认值 1。
 * 测试里用 c.compile_obj(src, out, ["-DYS_FACTOR=10"]) 编译，
 * 结果应当被放大 10 倍 —— 证明 -D 宏确实在预处理阶段生效了。 */
#ifndef YS_FACTOR
#define YS_FACTOR 1
#endif

int64_t ys_scaled(int64_t v) { return v * YS_FACTOR; }

int64_t ys_factor(void) { return YS_FACTOR; }

int64_t ys_add(int64_t a, int64_t b) { return a + b; }

int64_t ys_sub(int64_t a, int64_t b) { return a - b; }

int64_t ys_mul(int64_t a, int64_t b) { return a * b; }

int64_t ys_fact(int64_t n) {
    int64_t r = 1;
    for (int64_t k = 2; k <= n; k++) r *= k;
    return r;
}

/* 调用 libc 的 sqrt（需要链接 -lm，由 cgo 前置声明统一提供） */
double ys_sqrt(double x) { return sqrt(x); }

double ys_hypot2(double a, double b) { return sqrt(a * a + b * b); }

/* 数组演示：脚本侧不直接构造 int64 数组（语言层没有这个能力），
 * 而是让 C 侧填好缓冲，再用指针传回来读 —— 顺便演示 bytes ↔ 裸指针。 */
void ys_fill_i64(int64_t* buf, int64_t n) {
    for (int64_t i = 0; i < n; i++) buf[i] = i * i;
}

int64_t ys_sum_i64(const int64_t* buf, int64_t n) {
    int64_t s = 0;
    for (int64_t i = 0; i < n; i++) s += buf[i];
    return s;
}

int64_t ys_get_i64(const int64_t* buf, int64_t i) { return buf[i]; }

/* 就地翻倍缓冲，返回处理后的元素个数 */
int64_t ys_double_buf(int64_t* buf, int64_t n) {
    for (int64_t i = 0; i < n; i++) buf[i] *= 2;
    return n;
}
