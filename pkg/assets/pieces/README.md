# Piece sets

The piece sets used by image rendering, from Lichess. Only sets whose licences
allow them in a GPLv3 program are included (Lichess also has sets under
non-commercial licences, which aren't).

| Set | Author | Licence |
|---|---|---|
| cburnett | [Colin M.L. Burnett](https://en.wikipedia.org/wiki/User:Cburnett) | GPL-2.0-or-later |
| merida | Armando Hernandez Marroquin | GPL-2.0-or-later |
| mpchess | [Maxime Chupin](https://github.com/chupinmaxime) | GPL-3.0-or-later |
| chessnut | [Alexis Luengas](https://github.com/LexLuengas/chessnut-pieces) | Apache-2.0 |
| fantasy | [Maurizio Monge](https://github.com/maurimo/chess-art) | MIT |
| spatial | [Maurizio Monge](https://github.com/maurimo/chess-art) | MIT |
| celtic | [Maurizio Monge](https://github.com/maurimo/chess-art) | MIT |
| rhosgfx | [RhosGFX](https://rhosgfx.itch.io/) | CC0-1.0 |
| kiwen-suwi | [neverRare](https://github.com/neverRare) | CC-BY-4.0 |
| totoy | Kosal Sen | CC-BY-4.0 |
| papercut | [Nikolay Anzarov](https://nikoichu.itch.io/) | CC-BY-4.0 |
| shapes | [flugsio](https://github.com/flugsio/chess_shapes) | CC-BY-SA-4.0 |

Each folder has the original SVGs, PNGs rendered from them (embedded in the
program) and a LICENSE.md. To add a set, put its 12 SVGs (`wK.svg` … `bP.svg`)
in a new folder with a LICENSE.md, run `scripts/render-pieces.sh`, and add it
to `pieceSets` in `pkg/images.go`.
