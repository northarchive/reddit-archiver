package cmd

import (
	"github.com/northarchive/reddit-archiver/internal/downloader"
	"github.com/spf13/viper"

	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use: "start",
	//Short: "",
	//Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		downloader.Run()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	viper.SetDefault("output_dir", "./redditArchive")

	startCmd.Flags().StringP("out", "o", "./redditArchive", "Output directory")
	viper.BindPFlag("output_dir", startCmd.Flags().Lookup("out"))

	viper.SetDefault("subreddit_list_file", "./list.txt")

	startCmd.Flags().StringP("subreddit_list", "l", "./list.txt", "File with newline-seperated list of subreddits")
	viper.BindPFlag("subreddit_list_file", startCmd.Flags().Lookup("subreddit_list"))
}
