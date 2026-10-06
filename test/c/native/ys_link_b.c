/* ys_link_b.c —— 多编译单元演示：被依赖的一方
 *
 * 单独编译成 ys_link_b.o，**不参与** ys_link_a.o 的编译。
 * ys_link_a 里对 ys_link_b_value 的引用要到「链接」阶段才被解析，
 * 这正是分阶段编译的意义：改 B 只需重编 B.o 再重新链接。
 */
#include <stdint.h>

int64_t ys_link_b_value(void) { return 4242; }

int64_t ys_link_b_double(int64_t x) { return x * 2; }
