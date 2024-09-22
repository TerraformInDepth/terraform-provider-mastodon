package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPostResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccPostResourceConfig("First Test Post"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mastodon_post.test", "content", "First Test Post"),
					resource.TestCheckResourceAttr("mastodon_post.test", "visibility", "public"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "mastodon_post.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccPostResourceConfig("Post After Update"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mastodon_post.test", "content", "Post After Update"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccPostResourceConfig(content string) string {
	return fmt.Sprintf(`
resource "mastodon_post" "test" {
  content = %[1]q
}
`, content)
}
