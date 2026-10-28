class Magpie < Formula
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1788/magpie-cli-darwin-arm64"
      sha256 "decaa41717a8563b3792715bd4b02f68ccf64817ee5bdad8cdf489920793d87d"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1788/magpie-cli-darwin-amd64"
      sha256 "7a7236954e75027280f1509751f47a7202f3baea3fcdd96009fc390fc4d28262"
    end
  end
  on_linux do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1788/magpie-cli-linux-arm64"
      sha256 "85ab40eac93be5e921471fd1d17a2f13653b1ab1fa8742f4c37cef079eafea6b"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1788/magpie-cli-linux-amd64"
      sha256 "1707dc544831833e8dee14df58ab476a2145b516adadf9bbeb2421a1d85865fd"
    end
  end

  def install
    bin.install Dir["magpie-cli-*"].first => "magpie"
  end

  test do
    assert_match "magpie #{version}", shell_output("#{bin}/magpie --version")
  end
end
