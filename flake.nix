{
  description = "rombik — код → блок-схема алгоритму за ДСТУ 19.701-90";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        # nix develop — готове середовище без ручної возні з node.
        devShells.default = pkgs.mkShell {
          packages = [
            pkgs.nodejs_22 # рушій, веб і CLI — усе на Node 22
            pkgs.librsvg # rsvg-convert: SVG → PNG/PDF для CLI
          ];
          shellHook = ''
            echo "rombik dev — node $(node -v), npm $(npm -v)"
            echo "  npm install      # залежності (npm-воркспейси)"
            echo "  npm test         # тести рушія"
            echo "  npm run dev      # веб (SvelteKit)"
            echo "  npm run build:cli && ./dist/rombik.mjs examples/grade.py -o grade.svg"
          '';
        };
      }
    );
}
