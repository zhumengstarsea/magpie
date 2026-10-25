class Magpie < Formula
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1539/magpie-cli-darwin-arm64"
      sha256 "963011a32575aa796748e026794aec03b93850d4ca3da6d8e8849a879432c3ea"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1539/magpie-cli-darwin-amd64"
      sha256 "a30f2f3b8733882635d55d1b1fb32d72392cf108aac9aec797ebecf476712b68"
    end
  end
  on_linux do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1539/magpie-cli-linux-arm64"
      sha256 "a06f0bf9bc45f5921cbbf2d077f098ebd6ba5e314347e291dc060d7c05eab16c"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1539/magpie-cli-linux-amd64"
      sha256 "276aee547c294426781018990fd8b4de3ee317cb2d9d0cc5253c300f72e2de33"
    end
  end

  def install
    bin.install Dir["magpie-cli-*"].first => "magpie"
  end

  test do
    assert_match "magpie #{version}", shell_output("#{bin}/magpie --version")
  end
end
