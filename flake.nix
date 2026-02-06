{
  description = "A pokemon game with websockets";

  inputs = {
    default-pkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
  };

  outputs = { self, default-pkgs }: let

      # SUPPORTED SYSTEMS
      supportedSystems = [ "x86_64-linux" "x86_64-darwin" "aarch64-linux" "aarch64-darwin" ];
      
      forAllSystems = default-pkgs.lib.genAttrs supportedSystems;

      nixpkgsFor = system : pkgs : import pkgs {
        inherit system;
        config.allowUnfree = true;
      };
    
  in
  {
    devShells = forAllSystems ( system: 
      let 
        defaultPkgs = nixpkgsFor system default-pkgs ;
      in 
      {
        default = defaultPkgs.mkShell {
          packages = with defaultPkgs; [
            # task managers
            moon
            process-compose
            # - Go - main language
            go 
            gopls 
            gotools 
            air
            # - NODE - for running tests
            nodejs 
            pnpm 
            # - Process orchestration
            process-compose

            coreutils

            # SQL Generation
            sqlc
          ];
        };
      }
    ); 
  };
}