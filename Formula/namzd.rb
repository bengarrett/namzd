class Namzd < Formula
  desc "Quickly find files by name or extension"
  homepage "https://github.com/bengarrett/namzd"
  url "https://github.com/bengarrett/namzd/archive/refs/tags/v1.3.2.tar.gz"
  sha256 "93b3411bec40265d5e4c19e3f0e817f44634c79478756b1370675ef9de730532"
  version "1.3.2"
  license "GPL-3.0-only"

  @commit = "dc611f6de34c5af9e86248189cf244e50c56a2c9"
  @build_date = "2026-08-07T22:26:01+10:00"

  livecheck do
    url :stable
    strategy :github_latest
  end

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w -X main.version=#{version} -X main.commit=#{self.class.instance_variable_get('@commit')} -X main.date=#{self.class.instance_variable_get('@build_date')}")
  end

  test do
    assert_match "namzd", shell_output("#{bin}/namzd --version")
  end
end
