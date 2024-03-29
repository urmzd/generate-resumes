package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
	"github.com/urmzd/generate-resumes/pkg/compilers"
	"github.com/urmzd/generate-resumes/pkg/definition"
	"github.com/urmzd/generate-resumes/pkg/generators"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	"encoding/json"
)

var (
	ClassesFolder  string
	TemplateFolder string
	OutputsFolder  string
	KeepTex        bool
)

func initRootCmd() {
	rootCmd.PersistentFlags().StringVarP(&ClassesFolder, "classes", "c", "./assets/classes", "Define the style classes that can be used.")
	rootCmd.PersistentFlags().StringVarP(&TemplateFolder, "templates", "t", "./assets/templates", "Define the templates that can be used.")
	rootCmd.PersistentFlags().StringVarP(&OutputsFolder, "outputs", "o", "./outputs", "Defines the location to output the compiled files.")
	rootCmd.PersistentFlags().BoolVarP(&KeepTex, "keep-tex", "k", false, "Keep .tex files after compilation")
}

var rootCmd = &cobra.Command{
	Use:   "generate-resumes CONFIG",
	Short: "Generate beautiful LaTeX resumes with one command.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		logger, _ := zap.NewProduction()
		sugar := logger.Sugar()

		filename := args[0]

		data, err := os.ReadFile(filename)
		config := string(data)

		if err != nil {
			sugar.Fatalf("Error reading config file: %s", err)
		}

		var resume definition.Resume
		if strings.HasSuffix(filename, ".toml") {
			_, err = toml.Decode(config, &resume)
		} else if strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml") {
			decoder := yaml.NewDecoder(strings.NewReader(config))
			err = decoder.Decode(&resume)
		} else if strings.HasSuffix(filename, ".json") {
			decoder := json.NewDecoder(strings.NewReader(config))
			err = decoder.Decode(&resume)
		} else {
			panic("Unknown file type.")
		}

		if err != nil {
			panic(err)
		}

		compiler := compilers.NewXelatexCompiler("xelatex", sugar)
		compiler.AddOutputFolder(OutputsFolder)
		compiler.LoadClasses(ClassesFolder)

		generator := generators.NewBaseResumeGenerator(sugar)

		templateFiles, err := os.ReadDir(TemplateFolder)
		if err != nil {
			sugar.Fatal("Error reading template folder:", err)
		}
		for idx, file := range templateFiles {
			templatePath := filepath.Join(TemplateFolder, file.Name())
			tmpl, err := loadTemplate(templatePath)
			if err != nil {
				sugar.Error("Error loading template:", err)
				continue
			}

			latex := generator.Generate(tmpl, &resume)

			contactName := strings.ReplaceAll(resume.Contact.Name, " ", "_")
			templateType := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
			resumeFileName := fmt.Sprintf("%s_%d", contactName, idx)
			resumeFilePath := compiler.Compile(latex, resumeFileName)

			if KeepTex {
				cleanArtifacts(sugar, ".tex")
			} else {
				cleanArtifacts(sugar)
			}

			sugar.Infof("Generate resume of type %s at %s", templateType, resumeFilePath)
		}
	},
}

func loadTemplate(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	dataStr := string(data)
	return dataStr, err
}

func cleanArtifacts(logger *zap.SugaredLogger, keepExtensions ...string) {
	files, err := os.ReadDir(OutputsFolder)
	if err != nil {
		logger.Fatal("Error reading output folder:", err)
	}

	// Create a map for quick lookup of extensions to keep
	keep := make(map[string]bool)
	for _, ext := range keepExtensions {
		keep[ext] = true
	}
	// Always keep PDF files
	keep[".pdf"] = true

	for _, file := range files {
		if _, ok := keep[filepath.Ext(file.Name())]; !ok {
			// If the file's extension is not in the keep list, remove it
			err := os.Remove(filepath.Join(OutputsFolder, file.Name()))
			if err != nil {
				logger.Error("Error removing file:", err)
			}
		}
	}
}

func Execute() error {
	initRootCmd()
	return rootCmd.Execute()
}
