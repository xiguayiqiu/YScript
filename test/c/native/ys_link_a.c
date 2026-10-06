/* ys_link_a.c —— 多编译单元演示：依赖对方的一方
 *
 * ys_link_b_value / ys_link_b_double 声明在这里，但**定义**在另一个 .c 里。
 * 编译本文件时链接器不会介入（符号未定义是允许的），
 * 只有把 ys_link_a.o 与 ys_link_b.o 一起链接成 .so 时才补齐。
 */
#include <stdint.h>

/* 下面两个符号定义在 ys_link_b.c */
extern int64_t ys_link_b_value(void);
extern int64_t ys_link_b_double(int64_t x);

int64_t ys_link_a_value(void) { return ys_link_b_value() + 1; }

/* 跨编译单元的两级调用：a -> b -> back */
int64_t ys_link_chain(int64_t x) {
    return ys_link_b_double(ys_link_a_value()) + x;
}
