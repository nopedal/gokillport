# To use this formula, create a tap repo at github.com/nikolasopedal/homebrew-tap
# and place this file in Formula/gokillport.rb
#
# Users can then install with:
#   brew tap nikolasopedal/tap
#   brew install gokillport

class Gokillport < Formula
  desc "Find and kill processes listening on given ports"
  homepage "https://github.com/nopedal/gokillport"
  version "0.1.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/nopedal/gokillport/releases/download/v#{version}/gokillport-darwin-arm64"
      sha256 "PLACEHOLDER_SHA256_DARWIN_ARM64"
    else
      url "https://github.com/nopedal/gokillport/releases/download/v#{version}/gokillport-darwin-amd64"
      sha256 "PLACEHOLDER_SHA256_DARWIN_AMD64"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/nopedal/gokillport/releases/download/v#{version}/gokillport-linux-arm64"
      sha256 "PLACEHOLDER_SHA256_LINUX_ARM64"
    else
      url "https://github.com/nopedal/gokillport/releases/download/v#{version}/gokillport-linux-amd64"
      sha256 "PLACEHOLDER_SHA256_LINUX_AMD64"
    end
  end

  def install
    binary = Dir["gokillport*"].first
    bin.install binary => "gokillport"
    bin.install_symlink "gokillport" => "killport"
  end

  test do
    assert_match "gokillport version", shell_output("#{bin}/gokillport --version")
  end
end
