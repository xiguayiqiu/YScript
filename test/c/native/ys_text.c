/* ys_text.c —— 纯 C 字符串处理
 *
 * 演示点：
 *   - const char* 入参 ↔ YScript 字符串（cgo 侧自动做 NUL 结尾拷贝）
 *   - char* 出参：C 往调用方给的缓冲区里写（需要脚本先 c.malloc）
 *   - 返回 char*（动态分配）↔ YScript 字符串
 */
#include <stdint.h>
#include <string.h>
#include <stdlib.h>
#include <ctype.h>

/* 统计字符出现次数 */
int64_t ys_count_char(const char* s, char ch) {
    int64_t n = 0;
    for (; *s; s++) if (*s == ch) n++;
    return n;
}

/* 统计大写字母个数 */
int64_t ys_upper_count(const char* s) {
    int64_t n = 0;
    for (; *s; s++) if (isupper((unsigned char)*s)) n++;
    return n;
}

/* 就地反转：buf 需足够大，返回反转后的长度 */
int64_t ys_reverse(char* buf) {
    size_t n = strlen(buf);
    for (size_t i = 0; i < n / 2; i++) {
        char t = buf[i];
        buf[i] = buf[n - 1 - i];
        buf[n - 1 - i] = t;
    }
    return (int64_t)n;
}

/* 分配并返回一个新的全大写字符串（调用方需 c.free） */
char* ys_to_upper_dup(const char* s) {
    size_t n = strlen(s);
    char* out = (char*)malloc(n + 1);
    if (!out) return NULL;
    for (size_t i = 0; i < n; i++) out[i] = (char)toupper((unsigned char)s[i]);
    out[n] = '\0';
    return out;
}

/* 字节级异或：演示 bytes ↔ unsigned char* */
int64_t ys_xor_bytes(unsigned char* p, int64_t n, unsigned char key) {
    int64_t acc = 0;
    for (int64_t i = 0; i < n; i++) { p[i] ^= key; acc += p[i]; }
    return acc;
}
