#include <SDL3/SDL.h>
#include <stdio.h>
int main(void) {
    for (int sc = 1; sc < SDL_SCANCODE_COUNT; sc++) {
        SDL_Keycode k = SDL_GetKeyFromScancode((SDL_Scancode)sc, SDL_KMOD_NONE, false);
        if (k == SDLK_UNKNOWN) continue;
        const char *n = SDL_GetKeyName(k);
        if (!n || !*n) continue;
        printf("%u\t%d\t%s\t%s\n", (unsigned)k, sc, n, SDL_GetScancodeName((SDL_Scancode)sc));
    }
    return 0;
}
