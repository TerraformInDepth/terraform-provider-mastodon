provider "mastodon" {
  host = "https://mastodon.social"
}

resource "mastodon_post" "example" {
  content = "I'm posting to the Fediverse from Terraform!"
}
