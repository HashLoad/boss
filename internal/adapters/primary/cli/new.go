// Package cli provides command-line interface implementation for Boss.
package cli

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/pkg/consts"
	"github.com/hashload/boss/pkg/msg"
	"github.com/hashload/boss/pkg/pkgmanager"
	"github.com/spf13/cobra"
)

var (
	projectType string
	targetIDE   string
	quietNew    bool
)

const dprTemplate = `program %s;

{$APPTYPE CONSOLE}

{$R *.res}

uses
  System.SysUtils;

begin
  try
    Writeln('Hello from %s!');
  except
    on E: Exception do
      Writeln(E.ClassName, ': ', E.Message);
  end;
end.
`

const dpkTemplate = `package %s;

{$R *.res}
{$IFDEF IMPLICITBUILDING}
{$IMPLICITBUILD ON}
{$ENDIF}

requires
  rtl;

contains
  // Add package units here
  ;

end.
`

const dprojTemplate = `<Project xmlns="http://schemas.microsoft.com/developer/msbuild/2003">
    <PropertyGroup>
        <ProjectGuid>%s</ProjectGuid>
        <ProjectVersion>19.5</ProjectVersion>
        <FrameworkType>None</FrameworkType>
        <Base>True</Base>
        <Config Condition="'$(Config)'==''">Debug</Config>
        <Platform Condition="'$(Platform)'==''">Win32</Platform>
        <TargetedPlatforms>1</TargetedPlatforms>
        <AppType>%s</AppType>
    </PropertyGroup>
    <PropertyGroup Condition="'$(Config)'=='Base' or '$(Base)'!=''">
        <Base>true</Base>
    </PropertyGroup>
    <PropertyGroup Condition="('$(Platform)'=='Win32' and '$(Base)'=='true') or '$(Base_Win32)'!=''">
        <Base_Win32>true</Base_Win32>
        <CfgParent>Base</CfgParent>
        <Base>true</Base>
    </PropertyGroup>
    <PropertyGroup Condition="'$(Config)'=='Debug' or '$(Cfg_1)'!=''">
        <Cfg_1>true</Cfg_1>
        <CfgParent>Base</CfgParent>
        <Base>true</Base>
    </PropertyGroup>
    <PropertyGroup Condition="'$(Config)'=='Release' or '$(Cfg_2)'!=''">
        <Cfg_2>true</Cfg_2>
        <CfgParent>Base</CfgParent>
        <Base>true</Base>
    </PropertyGroup>
    <ItemGroup>
        <DelphiCompile Include="$(MainSource)">
            <MainSource>MainSource</MainSource>
        </DelphiCompile>
        <BuildConfiguration Include="Release">
            <Key>Cfg_2</Key>
            <CfgParent>Base</CfgParent>
        </BuildConfiguration>
        <BuildConfiguration Include="Base">
            <Key>Base</Key>
        </BuildConfiguration>
        <BuildConfiguration Include="Debug">
            <Key>Cfg_1</Key>
            <CfgParent>Base</CfgParent>
        </BuildConfiguration>
    </ItemGroup>
    <ProjectExtensions>
        <Borland.Personality>Delphi.Personality.12</Borland.Personality>
        <Borland.ProjectType>%s</Borland.ProjectType>
        <BorlandProject>
            <Delphi.Personality>
                <Source>
                    <Source Name="MainSource">%s.%s</Source>
                </Source>
            </Delphi.Personality>
        </BorlandProject>
        <ProjectFileVersion>12</ProjectFileVersion>
    </ProjectExtensions>
    <Import Project="$(BDS)\Bin\CodeGear.Delphi.Targets" Condition="Exists('$(BDS)\Bin\CodeGear.Delphi.Targets')"/>
    <Import Project="$(APPDATA)\Embarcadero\$(BDSAPPDATABASEDIR)\$(PRODUCTVERSION)\UserTools.proj" Condition="Exists('$(APPDATA)\Embarcadero\$(BDSAPPDATABASEDIR)\$(PRODUCTVERSION)\UserTools.proj')"/>
</Project>
`

const lprTemplate = `program %s;

{$mode objfpc}{$H+}

uses
  {$IFDEF UNIX}
  cthreads,
  {$ENDIF}
  Classes, SysUtils;

begin
  WriteLn('Hello from %s!');
end.
`

const lpiTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<CONFIG>
  <ProjectOptions>
    <Version Value="12"/>
    <General>
      <Flags>
        <MainUnitHasCreateFormStatements Value="False"/>
        <MainUnitHasTitleStatement Value="False"/>
        <MainUnitHasScaledStatement Value="False"/>
      </Flags>
      <SessionStorage Value="InProjectDir"/>
      <Title Value="%s"/>
      <UseAppBundle Value="False"/>
      <ResourceType Value="res"/>
    </General>
    <BuildModes>
      <Item Name="Default" Default="True"/>
    </BuildModes>
    <PublishOptions>
      <Version Value="2"/>
      <UseFileFilters Value="True"/>
    </PublishOptions>
    <RunParams>
      <FormatVersion Value="2"/>
    </RunParams>
    <Units>
      <Unit>
        <Filename Value="%s.lpr"/>
        <IsPartOfProject Value="True"/>
      </Unit>
    </Units>
  </ProjectOptions>
  <CompilerOptions>
    <Version Value="11"/>
    <Target>
      <Filename Value="%s"/>
    </Target>
    <SearchPaths>
      <IncludeFiles Value="$(ProjOutDir)"/>
      <UnitOutputDirectory Value="lib/$(TargetCPU)-$(TargetOS)"/>
    </SearchPaths>
  </CompilerOptions>
</CONFIG>
`

const lpkTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<CONFIG>
  <Package Version="5">
    <Name Value="%s"/>
    <Type Value="RunAndDesignTime"/>
    <CompilerOptions>
      <Version Value="11"/>
      <SearchPaths>
        <UnitOutputDirectory Value="lib/$(TargetCPU)-$(TargetOS)"/>
      </SearchPaths>
    </CompilerOptions>
    <Files>
      <!-- Add package files here -->
    </Files>
    <RequiredPkgs>
      <Item>
        <PackageName Value="FCL"/>
      </Item>
    </RequiredPkgs>
    <UsageOptions>
      <UnitPath Value="$(PkgOutDir)"/>
    </UsageOptions>
    <PublishOptions>
      <Version Value="2"/>
      <UseFileFilters Value="True"/>
    </PublishOptions>
  </Package>
</CONFIG>
`

// generateGUID generates a random GUID in the standard {XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX} format.
func generateGUID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "{00000000-0000-0000-0000-000000000000}"
	}
	return fmt.Sprintf("{%02X%02X%02X%02X-%02X%02X-%02X%02X-%02X%02X-%02X%02X%02X%02X%02X%02X}",
		b[0], b[1], b[2], b[3], b[4], b[5], b[6], b[7], b[8], b[9], b[10], b[11], b[12], b[13], b[14], b[15])
}

// newCmdRegister registers the new command.
func newCmdRegister(root *cobra.Command) {
	var newCmd = &cobra.Command{
		Use:     "new [project_name]",
		Short:   "Create a new Delphi or Lazarus project skeleton",
		Long:    "Create a new Delphi or Lazarus project skeleton with source directories, templates, and boss.json",
		Args:    cobra.MaximumNArgs(1),
		Example: `  Create a new console application (Delphi by default):
  boss new my_project

  Create a new Lazarus application:
  boss new my_project --ide lazarus

  Create a new package/library in Lazarus:
  boss new my_package --type pkg --ide lazarus`,
		Run: func(cmd *cobra.Command, args []string) {
			var name string
			if len(args) > 0 {
				name = args[0]
			}
			doCreateProject(name, projectType, targetIDE, quietNew)
		},
	}

	newCmd.Flags().StringVarP(&projectType, "type", "t", "app", "type of project to generate (app or pkg)")
	newCmd.Flags().StringVarP(&targetIDE, "ide", "i", "", "target IDE to generate for (delphi or lazarus)")
	newCmd.Flags().BoolVarP(&quietNew, "quiet", "q", false, "without asking questions")

	root.AddCommand(newCmd)
}

// doCreateProject performs the project creation.
func doCreateProject(name string, pType string, ide string, quiet bool) {
	if !quiet && name == "" {
		name = getParamOrDef("Project name", "")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		msg.Die("❌ Project name is required.")
	}

	pType = strings.ToLower(strings.TrimSpace(pType))
	if pType != "app" && pType != "pkg" {
		msg.Die("❌ Invalid project type. Supported types: 'app' (default) or 'pkg'.")
	}

	if !quiet && ide == "" {
		ide = getParamOrDef("Target IDE (delphi or lazarus)", "delphi")
	}
	ide = strings.ToLower(strings.TrimSpace(ide))
	if ide == "l" || ide == "lazarus" {
		ide = "lazarus"
	} else {
		ide = "delphi"
	}

	cwd, err := os.Getwd()
	if err != nil {
		msg.Die("❌ Failed to get current working directory: %v", err)
	}

	projectDir := filepath.Join(cwd, name)
	if _, err := os.Stat(projectDir); !os.IsNotExist(err) {
		msg.Die("❌ Directory '%s' already exists.", name)
	}

	ideTitle := "Delphi"
	if ide == "lazarus" {
		ideTitle = "Lazarus"
	}

	if !quiet {
		msg.Info("🚀 Creating a new %s project skeleton in %s...", ideTitle, projectDir)
	}

	// Create directories
	srcDir := filepath.Join(projectDir, "src")
	testsDir := filepath.Join(projectDir, "tests")
	if err := os.MkdirAll(srcDir, 0750); err != nil {
		msg.Die("❌ Failed to create src directory: %v", err)
	}
	if err := os.MkdirAll(testsDir, 0750); err != nil {
		msg.Die("❌ Failed to create tests directory: %v", err)
	}

	// Save boss.json
	packageData := domain.NewPackage()
	packageData.Name = name
	packageData.Version = "1.0.0"
	packageData.MainSrc = "src"

	packageJsonPath := filepath.Join(projectDir, consts.FilePackage)
	if err := pkgmanager.SavePackage(packageData, packageJsonPath); err != nil {
		msg.Die("❌ Failed to save boss.json: %v", err)
	}

	// Write files based on the chosen IDE
	if ide == "lazarus" {
		if pType == "app" {
			lprPath := filepath.Join(projectDir, name+".lpr")
			lprContent := fmt.Sprintf(lprTemplate, name, name)
			if err := os.WriteFile(lprPath, []byte(lprContent), 0644); err != nil {
				msg.Die("❌ Failed to create .lpr project file: %v", err)
			}

			lpiPath := filepath.Join(projectDir, name+".lpi")
			lpiContent := fmt.Sprintf(lpiTemplate, name, name, name)
			if err := os.WriteFile(lpiPath, []byte(lpiContent), 0644); err != nil {
				msg.Die("❌ Failed to create .lpi project file: %v", err)
			}
		} else {
			lpkPath := filepath.Join(projectDir, name+".lpk")
			lpkContent := fmt.Sprintf(lpkTemplate, name)
			if err := os.WriteFile(lpkPath, []byte(lpkContent), 0644); err != nil {
				msg.Die("❌ Failed to create .lpk package file: %v", err)
			}
		}
	} else {
		// Write Delphi files
		guid := generateGUID()
		var dprojContent string
		if pType == "app" {
			dprPath := filepath.Join(projectDir, name+".dpr")
			dprContent := fmt.Sprintf(dprTemplate, name, name)
			if err := os.WriteFile(dprPath, []byte(dprContent), 0644); err != nil {
				msg.Die("❌ Failed to create .dpr project file: %v", err)
			}
			dprojContent = fmt.Sprintf(dprojTemplate, guid, "Console", "Application", name, "dpr")
		} else {
			dpkPath := filepath.Join(projectDir, name+".dpk")
			dpkContent := fmt.Sprintf(dpkTemplate, name)
			if err := os.WriteFile(dpkPath, []byte(dpkContent), 0644); err != nil {
				msg.Die("❌ Failed to create .dpk package file: %v", err)
			}
			dprojContent = fmt.Sprintf(dprojTemplate, guid, "Package", "Package", name, "dpk")
		}

		dprojPath := filepath.Join(projectDir, name+".dproj")
		if err := os.WriteFile(dprojPath, []byte(dprojContent), 0644); err != nil {
			msg.Die("❌ Failed to create .dproj configuration file: %v", err)
		}
	}

	if !quiet {
		msg.Info("✨ Project '%s' created successfully!", name)
	}
}
