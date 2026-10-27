class Magpie < Formula
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1774/magpie-cli-darwin-arm64"
      sha256 "93ab0cfb4f6464ef0550003b309238fb94e77bc396c274857b5f15ff036825a9"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1774/magpie-cli-darwin-amd64"
      sha256 "1c3ae3cc4459ea303de06fb9f680e665c782431bb5d7960e4fb44b07daba2b77"
    end
  end
  on_linux do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1774/magpie-cli-linux-arm64"
      sha256 "3826900cf389c7b7b922ae7b6a6ca75486f97d1e1c78a49310e45fa1056ae568"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1774/magpie-cli-linux-amd64"
      sha256 "11823d766072d4b3602af53c9a3aefc64eee2dbf636e0148660d1fc77becc0a8"
    end
  end

  def install
    bin.install Dir["magpie-cli-*"].first => "magpie"
  end

  test do
    assert_match "magpie #{version}", shell_output("#{bin}/magpie --version")
  end
end
