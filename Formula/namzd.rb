class Namzd < Formula
  desc "Quickly find files by name or extension"
  homepage "https://github.com/bengarrett/namzd"
  url "https://github.com/bengarrett/namzd/archive/refs/tags/v1.2.4.tar.gz"
  sha256 "29c25beebb1b69037c6877332181fda8e1d0c60a0f4390476f155bde7f652bf6"
  version "1.2.4"
  license "GPL-3.0-only"

  livecheck do
    url :stable
    strategy :github_latest
  end

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w")
  end

  test do
    assert_match "namzd", shell_output("#{bin}/namzd --version")
  end
end
