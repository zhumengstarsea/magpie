class Magpie < Formula
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1734/magpie-cli-darwin-arm64"
      sha256 "f090b3ea45086642cd1d17d2b46ffed8507f6fff2d45fed52502c0a4bfe0838c"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1734/magpie-cli-darwin-amd64"
      sha256 "4d4c9f315e30c96cd0b26624bd712b3db5252842f2839f48e188708af74f66bd"
    end
  end
  on_linux do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1734/magpie-cli-linux-arm64"
      sha256 "26310597f75a970fc966018c8084a2334f79a7c33f9fb9e502c8a79ed14eb8c9"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1734/magpie-cli-linux-amd64"
      sha256 "b83114f3c1bfee9d340fc75c755e668e7c96fde921e6cc20ff545ebbcfca39eb"
    end
  end

  def install
    bin.install Dir["magpie-cli-*"].first => "magpie"
  end

  test do
    assert_match "magpie #{version}", shell_output("#{bin}/magpie --version")
  end
end
