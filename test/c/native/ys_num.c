/* ys_num.c —— 数值类型互操作（libffi 后端）
 *
 * 演示点：
 *   - 浮点参数与浮点返回值（旧垫片后端只能按 int64 传，跑不出正确结果）
 *   - 混合 int64/double 参数
 *   - float(32 位) 与 double(64 位) 的区别
 *   - 窄整数类型（int16/uint8/int32）
 *   - 超过 8 个参数（int64 垫片后端的上限）
 */
#include <stdint.h>

double ys_mix(int64_t a, double b, int64_t c, double d) { return a + b + c + d; }

double ys_dsum(double a, double b, double c) { return a + b + c; }

float  ys_f32(float x) { return x * 2.0f; }

int16_t ys_i16(int16_t v) { return (int16_t)(v * 2); }

uint8_t ys_u8(uint8_t v) { return (uint8_t)(v + 1); }

int32_t ys_i32abs(int32_t v) { return v < 0 ? -v : v; }

int64_t ys_many12(int64_t a, int64_t b, int64_t c, int64_t d, int64_t e, int64_t f,
                  int64_t g, int64_t h, int64_t i, int64_t j, int64_t k, int64_t l) {
    return a+b+c+d+e+f+g+h+i+j+k+l;
}

/* 12 个参数里混一个浮点 */
double ys_many_mix(int64_t a, double b, int64_t c, int64_t d, int64_t e, int64_t f,
                   int64_t g, int64_t h, int64_t i, int64_t j, int64_t k, double l) {
    return (double)(a+c+e+g+i+k) + b + d + f + h + j + l;
}
