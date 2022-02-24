package cmd

import (
	"github.com/northarchive/reddit-archiver/internal/downloader"
	"github.com/spf13/viper"
	"os"

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

	viper.SetDefault("output_dir", "."+string(os.PathSeparator)+"redditArchive")

	startCmd.Flags().StringP("out", "o", "."+string(os.PathSeparator)+"redditArchive", "Output directory")
	viper.BindPFlag("output_dir", startCmd.Flags().Lookup("out"))

	viper.SetDefault("subreddit_list_file", "."+string(os.PathSeparator)+"list.txt")

	startCmd.Flags().StringP("subreddit_list", "l", "."+string(os.PathSeparator)+"list.txt", "File with newline-seperated list of subreddits")
	viper.BindPFlag("subreddit_list_file", startCmd.Flags().Lookup("subreddit_list"))
}
