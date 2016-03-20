.main
  ; set stack pointer
  ld sp, $dfff

  ; disable interrupts, set up for vblank
  di
  ld a, $01
  ldh ($ff), a

.wait_for_vblank
  ldh a, ($44)
  cp $94
  jr nz, wait_for_vblank

  ; disable display
  ld a, $00
  ldh ($40), a

  ; setup palettes
  ld a, $e4
  ldh ($47), a
  ldh ($48), a
  ldh ($49), a

  ; fill TILE_MAP_0 and TILE_DATA_0 with stuff
  ld a, $00
  ld hl, $9000
  ld de, $9800
.copy_tiles_outer
  ld b, $10
.copy_tiles_inner
  ldi (hl), a
  dec b
  jr nz, copy_tiles_inner
  ld (de), a
  inc de
  dec a
  jr nz, copy_tiles_outer

  ; enable display
  ld a, $81
  ldh ($40), a
  halt
