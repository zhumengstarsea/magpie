class Magpie < Formula
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1982/magpie-cli-darwin-arm64"
      sha256 "3f244369d331d74146666630ed334321cb51ec860867b86c1c25063e5c72d14b"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1982/magpie-cli-darwin-amd64"
      sha256 "7809aa87f1df414606bfbd5b99a9d418cc2a12342e3a97c1614bcec83051e5a4"
    end
  end
  on_linux do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1982/magpie-cli-linux-arm64"
      sha256 "2a83e8993b24d534f467391c263ff4256ca41c2da072f315e2a5d46cfd3dea05"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1982/magpie-cli-linux-amd64"
      sha256 "d53eb9242beaee514c6e77e39bda07ea643e6dad065d604c29f05bd8b9915af7"
    end
  end

  def install
    bin.install Dir["magpie-cli-*"].first => "magpie"
  end

  test do
    assert_match "magpie #{version}", shell_output("#{bin}/magpie --version")
  end
end
