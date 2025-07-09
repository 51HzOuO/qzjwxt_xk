package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// CourseSession represents a course selection session
type CourseSession struct {
	Term string // 学年学期
	Name string // 选课名称
	Time string // 选课时间
	URL  string // 选课URL
}

// Course represents a course from the response
type Course struct {
	Kch      string     `json:"kch"`      // 课程编号
	Kcmc     string     `json:"kcmc"`     // 课程名称
	Xf       int        `json:"xf"`       // 学分
	Skls     string     `json:"skls"`     // 上课老师
	Sksj     string     `json:"sksj"`     // 上课时间
	Skdd     string     `json:"skdd"`     // 上课地点
	Xqmc     string     `json:"xqmc"`     // 上课校区
	Syrs     string     `json:"syrs"`     // 剩余量
	Jx0404id string     `json:"jx0404id"` // 选课ID
	Szkcflmc string     `json:"szkcflmc"` // 通选课类别
	KkapList []KkapInfo `json:"kkapList"` // 课程安排信息
	Fzmc     string     `json:"fzmc"`     // 课程分组名称
}

// KkapInfo represents course arrangement information
type KkapInfo struct {
	Jgxm     string   `json:"jgxm"`     // 教师姓名
	Kkzc     string   `json:"kkzc"`     // 开课周次
	Xq       string   `json:"xq"`       // 星期
	Skjcmc   string   `json:"skjcmc"`   // 上课节次
	Jsmc     string   `json:"jsmc"`     // 教室名称
	SkzcList []string `json:"skzcList"` // 上课周次列表
}

// CourseResponse represents the JSON response structure
type CourseResponse struct {
	AaData []Course `json:"aaData"`
}

// Global variables
var courseMap map[string]string
var courseSections map[string][]string // Maps course number (kch) to all section IDs (jx0404id)
var selectedSession CourseSession      // Store the selected session globally

func main() {
	// Display disclaimer at startup
	fmt.Println("==============================================================================")
	fmt.Println("⚠️  警告：请勿使用该项目进行任何形式的商业盈利行为，包括但不限于收费服务、转售代码、嵌入付费软件等。")
	fmt.Println()
	fmt.Println("本项目旨在提供便捷的选课辅助工具，仅供学习与个人使用。")
	fmt.Println()
	fmt.Println("如需在公开平台分发、修改或复用本项目，请确保遵守GPL 协议条款，并注明原作者及项目来源。")
	fmt.Println()
	fmt.Println("感谢你的理解与支持。如果你有建议或改进意见，欢迎通过 Issue 或 Pull Request 的方式进行交流。")
	fmt.Println()
	fmt.Println("GitHub repo:")
	fmt.Println("https://github.com/51HzOuO/qzjwxt_xk")
	fmt.Println("==============================================================================")
	fmt.Println()

	// Step 1: Get username and password from user input
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("请输入账号: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Print("请输入密码: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	// Encode username and password in base64
	usernameBase64 := base64.StdEncoding.EncodeToString([]byte(username))
	passwordBase64 := base64.StdEncoding.EncodeToString([]byte(password))

	// Format exactly as in the example: MjAyMzEyMDA5Nzc4%25%25%25TGl1MDUwNDIw%3D
	encoded := fmt.Sprintf("%s%%25%%25%%25%s%%3D", usernameBase64, passwordBase64)

	fmt.Println("编码后的登录参数:", encoded)

	// Step 2: Login and get cookies
	cookies, err := login(encoded)
	if err != nil {
		fmt.Printf("登录失败: %v\n", err)
		fmt.Println("按任意键退出...")
		reader.ReadString('\n')
		return
	}

	fmt.Println("登录成功!")

	// Step 3: Request initial authentication and select course session
	fmt.Println("\n开始选课会话认证...")
	err = authenticate(cookies)
	if err != nil {
		fmt.Printf("认证失败: %v\n", err)
		return
	}

	// Step 4: Get course list
	fmt.Println("\n获取课程列表...")
	var getCourseErr error
	courseMap, getCourseErr = getCourseList(cookies)
	if getCourseErr != nil {
		fmt.Printf("获取课程列表失败: %v\n", getCourseErr)
		return
	}

	// Step 5: Let user select courses
	selectedCourses := selectCourses(courseMap)

	// Step 6: Register for selected courses
	fmt.Println("\n开始选课，将在每次尝试前自动刷新认证会话...")
	registerForCourses(selectedCourses, cookies, encoded)
}

// login sends a login request and returns cookies
func login(encoded string) ([]*http.Cookie, error) {
	// Create POST request with encoded parameter
	data := "encoded=" + encoded
	fmt.Println("发送的完整请求体:", data)

	req, err := http.NewRequest("POST", "https://jw.educationgroup.cn/ytkjxy_jsxsd/xk/LoginToXk",
		strings.NewReader(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Host", "jw.educationgroup.cn")
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(data)))

	// Print request details
	fmt.Println("\n请求详情:")
	fmt.Println("URL:", req.URL.String())
	fmt.Println("Method:", req.Method)
	fmt.Println("Headers:")
	for name, values := range req.Header {
		for _, value := range values {
			fmt.Printf("  %s: %s\n", name, value)
		}
	}

	// Disable automatic redirects to capture the 302 response
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Print response status for debugging
	fmt.Println("\n响应状态码:", resp.StatusCode)

	// For successful login, status should be 302 (redirect)
	if resp.StatusCode != 302 {
		// If we got 200, it means there was an error (login page with error message)
		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)

		fmt.Println("\n登录失败! 响应体预览:")
		previewLen := 500
		if len(body) < previewLen {
			previewLen = len(body)
		}
		fmt.Printf("%s\n", body[:previewLen])

		// Try to extract more specific error messages
		errorMsg := "登录失败"
		if strings.Contains(bodyStr, "密码错误") || strings.Contains(bodyStr, "密码不正确") {
			errorMsg = "密码错误"
		} else if strings.Contains(bodyStr, "账号不存在") || strings.Contains(bodyStr, "用户名不存在") {
			errorMsg = "账号不存在"
		} else if strings.Contains(bodyStr, "验证码") && strings.Contains(bodyStr, "错误") {
			errorMsg = "验证码错误"
		}

		return nil, fmt.Errorf("%s", errorMsg)
	}

	// Print all headers for debugging
	fmt.Println("响应头:")
	for name, values := range resp.Header {
		for _, value := range values {
			fmt.Printf("%s: %s\n", name, value)
		}
	}

	// Print all cookies
	cookies := resp.Cookies()
	fmt.Println("\n收到的Cookie:")
	for i, cookie := range cookies {
		fmt.Printf("%d. %s = %s (Domain: %s, Path: %s)\n",
			i+1, cookie.Name, cookie.Value, cookie.Domain, cookie.Path)
	}

	// Check if we have the necessary cookies
	if len(cookies) == 0 {
		return nil, fmt.Errorf("登录成功但未收到Cookie")
	}

	// Check for the redirect location
	location := resp.Header.Get("Location")
	if location != "" {
		fmt.Println("\n重定向地址:", location)
	}

	return cookies, nil
}

// authenticate sends an authentication request
func authenticate(cookies []*http.Cookie) error {
	// First, get the list of available course selection sessions
	sessions, err := getSessionList(cookies)
	if err != nil {
		return fmt.Errorf("failed to get session list: %v", err)
	}

	// Display available sessions to the user
	fmt.Println("\n可用的选课会话:")
	fmt.Printf("%-4s %-15s %-20s %-25s\n", "序号", "学年学期", "选课名称", "选课时间")
	fmt.Println(strings.Repeat("-", 70))

	for i, session := range sessions {
		fmt.Printf("%-4d %-15s %-20s %-25s\n", i+1, session.Term, session.Name, session.Time)
	}

	// Let user select a session
	reader := bufio.NewReader(os.Stdin)
	var localSelectedSession CourseSession

	// Always require manual selection, even if there's only one option
	for {
		fmt.Print("\n请选择选课会话编号: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		// Convert input to integer
		var sessionIndex int
		_, err := fmt.Sscanf(input, "%d", &sessionIndex)
		if err != nil || sessionIndex < 1 || sessionIndex > len(sessions) {
			fmt.Printf("无效的选择，请输入 1-%d 之间的数字\n", len(sessions))
			continue
		}

		localSelectedSession = sessions[sessionIndex-1]
		break
	}

	// Store the selected session in the global variable
	selectedSession = localSelectedSession

	fmt.Printf("\n已选择: %s - %s\n", selectedSession.Term, selectedSession.Name)
	fmt.Printf("使用URL: %s\n", selectedSession.URL)

	// Send authentication request with the selected session URL
	err = refreshAuthentication(cookies)
	if err != nil {
		return err
	}

	fmt.Println("认证成功!")
	return nil
}

// refreshAuthentication re-authenticates with the selected session URL
func refreshAuthentication(cookies []*http.Cookie) error {
	if selectedSession.URL == "" {
		return fmt.Errorf("没有选择选课会话")
	}

	authURL := "https://jw.educationgroup.cn" + selectedSession.URL
	req, err := http.NewRequest("GET", authURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Host", "jw.educationgroup.cn")

	// Add cookies to request
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		if len(body) > 0 {
			previewLen := 200
			if len(body) < previewLen {
				previewLen = len(body)
			}
			fmt.Printf("认证响应预览: %s\n", body[:previewLen])
		}
		return fmt.Errorf("认证失败，状态码: %d", resp.StatusCode)
	}

	// Check if the response contains indicators of successful authentication
	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	if strings.Contains(bodyStr, "权限不足") || strings.Contains(bodyStr, "请重新登录") {
		return fmt.Errorf("认证失败: 权限不足或会话已过期，请重新登录")
	}

	return nil
}

// getCourseList fetches the list of available courses
func getCourseList(cookies []*http.Cookie) (map[string]string, error) {
	// Create a map to store unique courses by jx0404id
	allCourses := make(map[string]Course)      // jx0404id -> Course
	coursesByKch := make(map[string][]Course)  // kch -> []Course
	courseMap := make(map[string]string)       // jx0404id -> kch (reversed from before)
	courseSections = make(map[string][]string) // kch -> []jx0404id

	// Create a WaitGroup to synchronize the goroutines
	var wg sync.WaitGroup
	var mutex sync.Mutex

	// Create a slice to collect errors
	var errors []string
	var errorsMutex sync.Mutex

	fmt.Println("正在获取所有星期的课程数据...")

	// Iterate through all days of the week (1-7)
	for day := 1; day <= 7; day++ {
		wg.Add(1)
		go func(skxq int) {
			defer wg.Done()

			fmt.Printf("获取星期 %d 的课程...\n", skxq)

			// Prepare the request data
			data := "sEcho=1&iColumns=12&sColumns=&iDisplayStart=0&iDisplayLength=9999" +
				"&mDataProp_0=kch&mDataProp_1=kcmc&mDataProp_2=fzmc&mDataProp_3=xf" +
				"&mDataProp_4=skls&mDataProp_5=sksj&mDataProp_6=skdd&mDataProp_7=xqmc" +
				"&mDataProp_8=xkrs&mDataProp_9=syrs&mDataProp_10=ctsm&mDataProp_11=czOper"

			// Create the URL with the specific skxq parameter
			url := fmt.Sprintf("https://jw.educationgroup.cn/ytkjxy_jsxsd/xsxkkc/xsxkFawxk?kcxx=&skls=&skxq=%d&skjc=&sfym=false&sfct=false&sfxx=true&skxq_xx0103=&kzyxkbx=0&kzyxkxx=0&kzyxkrx=0&kzyxkqt=0", skxq)

			req, err := http.NewRequest("POST", url, strings.NewReader(data))
			if err != nil {
				errorsMutex.Lock()
				errors = append(errors, fmt.Sprintf("星期 %d 请求创建失败: %v", skxq, err))
				errorsMutex.Unlock()
				return
			}

			req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
			req.Header.Set("Host", "jw.educationgroup.cn")

			// Add cookies to request
			for _, cookie := range cookies {
				req.AddCookie(cookie)
			}

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				errorsMutex.Lock()
				errors = append(errors, fmt.Sprintf("星期 %d 请求发送失败: %v", skxq, err))
				errorsMutex.Unlock()
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				errorsMutex.Lock()
				errors = append(errors, fmt.Sprintf("星期 %d 获取课程列表失败，状态码: %d", skxq, resp.StatusCode))
				errorsMutex.Unlock()
				return
			}

			// Read and parse the response
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				errorsMutex.Lock()
				errors = append(errors, fmt.Sprintf("星期 %d 读取响应失败: %v", skxq, err))
				errorsMutex.Unlock()
				return
			}

			// Check if response is HTML instead of JSON
			if strings.Contains(string(body), "<html") {
				errorsMutex.Lock()
				errors = append(errors, fmt.Sprintf("星期 %d 收到HTML响应而非JSON，会话可能已过期或认证失败", skxq))
				errorsMutex.Unlock()
				return
			}

			var courseResp CourseResponse
			err = json.Unmarshal(body, &courseResp)
			if err != nil {
				errorsMutex.Lock()
				errors = append(errors, fmt.Sprintf("星期 %d 解析JSON失败: %v", skxq, err))
				errorsMutex.Unlock()
				return
			}

			// Lock the mutex before updating the shared maps
			mutex.Lock()
			defer mutex.Unlock()

			// Add courses to the maps
			for _, course := range courseResp.AaData {
				// Store by jx0404id to avoid duplicates
				if _, exists := allCourses[course.Jx0404id]; !exists {
					allCourses[course.Jx0404id] = course

					// Group courses by kch
					coursesByKch[course.Kch] = append(coursesByKch[course.Kch], course)

					// Reverse mapping: jx0404id -> kch
					courseMap[course.Jx0404id] = course.Kch
				}
			}

			fmt.Printf("星期 %d 获取到 %d 门课程\n", skxq, len(courseResp.AaData))

		}(day)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Check if there were any errors
	if len(errors) > 0 {
		// If all days failed, return an error
		if len(errors) == 7 {
			return nil, fmt.Errorf("获取课程列表失败: %s", strings.Join(errors, "; "))
		}

		// Otherwise, just print the errors but continue
		fmt.Println("\n获取课程时遇到以下错误:")
		for _, err := range errors {
			fmt.Printf("- %s\n", err)
		}
		fmt.Println("但仍然获取到了部分课程数据，将继续处理...")
	}

	// If we didn't get any courses, return an error
	if len(allCourses) == 0 {
		return nil, fmt.Errorf("未能获取到任何课程数据")
	}

	// Populate the courseSections map
	for kch, courses := range coursesByKch {
		var sectionIDs []string
		for _, course := range courses {
			sectionIDs = append(sectionIDs, course.Jx0404id)
		}
		courseSections[kch] = sectionIDs
	}

	// Print table header
	fmt.Printf("\n共获取到 %d 门可选课程:\n", len(allCourses))
	fmt.Printf("%-15s %-20s %-4s %-10s %-20s %-20s %-8s %-6s %-20s %-10s\n",
		"选课ID", "课程名称", "学分", "教师", "上课时间", "上课地点", "上课校区", "剩余量", "通选课类别", "课程编号")
	fmt.Println(strings.Repeat("-", 140))

	// Print each course grouped by course number
	for _, courses := range coursesByKch {
		for _, course := range courses {
			// Get teacher name
			teacherName := course.Skls
			if len(course.KkapList) > 0 && course.KkapList[0].Jgxm != "" {
				teacherName = course.KkapList[0].Jgxm
			}

			// Get classroom
			classroom := course.Skdd
			if len(course.KkapList) > 0 && course.KkapList[0].Jsmc != "" {
				classroom = course.KkapList[0].Jsmc
			}

			// Format course time
			courseTime := course.Sksj
			if courseTime == "" && len(course.KkapList) > 0 {
				xq := ""
				switch course.KkapList[0].Xq {
				case "1":
					xq = "星期一"
				case "2":
					xq = "星期二"
				case "3":
					xq = "星期三"
				case "4":
					xq = "星期四"
				case "5":
					xq = "星期五"
				case "6":
					xq = "星期六"
				case "7":
					xq = "星期日"
				}
				courseTime = fmt.Sprintf("%s %s %s", course.KkapList[0].Kkzc, xq, course.KkapList[0].Skjcmc)
			}

			// Format remaining spots
			remainingSpots := course.Syrs
			if remainingSpots == "0" {
				remainingSpots = "满"
			}

			// Print course info in a formatted way
			fmt.Printf("%-15s %-20.20s %-4d %-10.10s %-20.20s %-20.20s %-8.8s %-6s %-20.20s %-10s\n",
				course.Jx0404id, course.Kcmc, course.Xf, teacherName, courseTime, classroom, course.Xqmc, remainingSpots, course.Szkcflmc, course.Kch)
		}
		// Add a separator between different course numbers
		fmt.Println(strings.Repeat("-", 140))
	}

	return courseMap, nil
}

// selectCourses lets the user select courses to register for
func selectCourses(courseMap map[string]string) []string {
	reader := bufio.NewReader(os.Stdin)

	// Use a map to track unique course selections
	selectedCoursesMap := make(map[string]struct{})

	fmt.Println("\n请输入要选择的选课ID，每行一个，输入 'done' 结束:")

	for {
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "done" {
			break
		}

		if _, exists := courseMap[input]; exists {
			if _, alreadySelected := selectedCoursesMap[input]; alreadySelected {
				fmt.Printf("选课ID %s 已经添加过了，请勿重复添加\n", input)
			} else {
				selectedCoursesMap[input] = struct{}{}
				fmt.Printf("已添加选课ID: %s\n", input)
			}
		} else {
			fmt.Printf("选课ID %s 不存在，请重新输入\n", input)
		}
	}

	// Convert map keys to slice
	var selectedCourses []string
	for jx0404id := range selectedCoursesMap {
		selectedCourses = append(selectedCourses, jx0404id)
	}

	fmt.Printf("\n已选择 %d 门课程\n", len(selectedCourses))
	return selectedCourses
}

// registerForCourses registers for the selected courses
func registerForCourses(selectedCourses []string, cookies []*http.Cookie, encoded string) {
	var wg sync.WaitGroup
	successChan := make(chan string)
	doneChan := make(chan bool)

	// Create a channel for token refresh requests
	tokenRefreshChan := make(chan bool)
	tokenRefreshDoneChan := make(chan []*http.Cookie)
	quitRefreshChan := make(chan bool) // Add a quit channel

	// Create a mutex to protect shared cookies
	var cookiesMutex sync.Mutex
	sharedCookies := cookies

	// Create an atomic flag to prevent multiple refresh requests
	// When a goroutine detects token expiration, it checks this flag
	// Only one goroutine will trigger the refresh process at a time
	var isRefreshing int32

	// Create a function to get the current cookies
	getCookies := func() []*http.Cookie {
		cookiesMutex.Lock()
		defer cookiesMutex.Unlock()
		return sharedCookies
	}

	// Create a function to update the shared cookies
	updateCookies := func(newCookies []*http.Cookie) {
		cookiesMutex.Lock()
		defer cookiesMutex.Unlock()
		sharedCookies = newCookies
	}

	// 使用传入的encoded参数，不再要求用户重新输入账号密码
	fmt.Println("\n将使用之前的登录信息自动处理会话过期问题")

	// Start a goroutine to handle token refresh
	go func() {
		for {
			select {
			case <-tokenRefreshChan:
				// Set the refreshing flag
				atomic.StoreInt32(&isRefreshing, 1)

				fmt.Println("\n检测到会话已过期，正在重新登录...")

				// Re-login
				newCookies, err := login(encoded)
				if err != nil {
					fmt.Printf("重新登录失败: %v\n", err)
					tokenRefreshDoneChan <- getCookies() // Return current cookies if login fails
					atomic.StoreInt32(&isRefreshing, 0)  // Reset the refreshing flag
					continue
				}

				fmt.Println("重新登录成功，正在刷新认证会话...")

				// Re-authenticate with the selected session
				err = refreshAuthentication(newCookies)
				if err != nil {
					fmt.Printf("刷新认证失败: %v\n", err)
					tokenRefreshDoneChan <- getCookies() // Return current cookies if authentication fails
					atomic.StoreInt32(&isRefreshing, 0)  // Reset the refreshing flag
					continue
				}

				fmt.Println("认证会话刷新成功，更新共享会话令牌...")
				updateCookies(newCookies)
				tokenRefreshDoneChan <- newCookies

				// Reset the refreshing flag
				atomic.StoreInt32(&isRefreshing, 0)

			case <-quitRefreshChan:
				// Exit the goroutine when quit signal is received
				return
			}
		}
	}()

	// Start a goroutine to collect successful registrations
	go func() {
		successfulCourses := []string{}
		for {
			select {
			case course := <-successChan:
				successfulCourses = append(successfulCourses, course)
				fmt.Printf("课程 %s 选课成功!\n", course)
			case <-doneChan:
				fmt.Println("\n选课结果汇总:")
				if len(successfulCourses) > 0 {
					fmt.Println("成功选上的课程:")
					for _, course := range successfulCourses {
						fmt.Printf("- %s\n", course)
					}
				} else {
					fmt.Println("没有成功选上任何课程")
				}
				// Signal the token refresh goroutine to exit
				quitRefreshChan <- true
				return
			}
		}
	}()

	// Start a goroutine for each course
	for _, jx0404id := range selectedCourses {
		wg.Add(1)
		go func(jx0404id string) {
			defer wg.Done()

			// Get the course number for display purposes
			kch := courseMap[jx0404id]

			attempts := 0

			// Continue indefinitely until successful or manually stopped
			for {
				attempts++

				// Get the latest cookies
				localCookies := getCookies()

				// Use the new API endpoint for course selection
				url := fmt.Sprintf("https://jw.educationgroup.cn/ytkjxy_jsxsd/xsxkkc/fawxkOper?jx0404id=%s&xkzy=&trjf=&_=%d",
					jx0404id, time.Now().UnixMilli())

				req, err := http.NewRequest("GET", url, nil)
				if err != nil {
					fmt.Printf("课程 %s 请求创建失败: %v\n", jx0404id, err)
					time.Sleep(1 * time.Second)
					continue
				}

				req.Header.Set("Host", "jw.educationgroup.cn")

				// Add cookies to request
				for _, cookie := range localCookies {
					req.AddCookie(cookie)
				}

				client := &http.Client{}
				resp, err := client.Do(req)
				if err != nil {
					fmt.Printf("课程 %s 请求发送失败: %v\n", jx0404id, err)
					time.Sleep(1 * time.Second)
					continue
				}

				body, err := io.ReadAll(resp.Body)
				resp.Body.Close()
				if err != nil {
					fmt.Printf("课程 %s 响应读取失败: %v\n", jx0404id, err)
					time.Sleep(1 * time.Second)
					continue
				}

				bodyStr := string(body)

				// 检查是否需要重新登录 - 两种情况:
				// 1. 响应是HTML而不是JSON (表示token过期)
				// 2. 响应是JSON但包含"当前账号已在别处登录"信息
				needRelogin := strings.Contains(bodyStr, "<html") ||
					strings.Contains(bodyStr, "当前账号已在别处登录") ||
					strings.Contains(bodyStr, "请重新登录")

				if needRelogin {
					var reason string
					if strings.Contains(bodyStr, "<html") {
						reason = "收到HTML响应"
					} else {
						reason = "账号在别处登录或会话已过期"
					}

					fmt.Printf("课程 %s %s，会话可能已过期\n", jx0404id, reason)

					// Check if a refresh is already in progress
					// CompareAndSwap atomically sets isRefreshing to 1 if it was 0 and returns true
					// This ensures only one goroutine will trigger the refresh process
					if atomic.CompareAndSwapInt32(&isRefreshing, 0, 1) {
						// We're the first to trigger a refresh
						fmt.Printf("课程 %s 正在触发会话刷新...\n", jx0404id)

						// Reset the flag (the refresh goroutine will set it properly)
						atomic.StoreInt32(&isRefreshing, 0)

						// Request token refresh
						tokenRefreshChan <- true
					} else {
						fmt.Printf("课程 %s 等待会话刷新完成...\n", jx0404id)
					}

					// All goroutines that detect token expiration will wait here
					// They will each receive the new cookies when refresh is complete
					newCookies := <-tokenRefreshDoneChan

					// Explicitly use the new cookies for the next request
					localCookies = newCookies

					fmt.Printf("课程 %s 已获取新的会话令牌，继续选课...\n", jx0404id)
					continue
				}

				// Print response for debugging (without HTML content)
				if !strings.Contains(bodyStr, "<html") {
					fmt.Printf("课程 %s 响应: %s\n", jx0404id, bodyStr)
				} else {
					fmt.Printf("课程 %s 收到HTML响应 (内容已省略)\n", jx0404id)
				}

				// Parse the response
				var result struct {
					Success interface{} `json:"success"`
					Message string      `json:"message"`
				}

				err = json.Unmarshal(body, &result)
				if err != nil {
					fmt.Printf("课程 %s 响应解析失败: %v\n", jx0404id, err)
					time.Sleep(1 * time.Second)
					continue
				}

				// 处理success字段可能是布尔值或数组的情况
				var isSuccess bool
				switch v := result.Success.(type) {
				case bool:
					isSuccess = v
				case []interface{}:
					// 如果是数组，检查第一个元素是否为true
					if len(v) > 0 {
						if b, ok := v[0].(bool); ok {
							isSuccess = b
						}
					}
				default:
					isSuccess = false
				}

				if isSuccess && result.Message == "选课成功" {
					successChan <- fmt.Sprintf("%s (课程编号: %s)", jx0404id, kch)
					return
				}

				fmt.Printf("课程 %s 尝试 %d: %s\n", jx0404id, attempts, result.Message)
				time.Sleep(1 * time.Second)
			}
		}(jx0404id)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	doneChan <- true
}

// getSessionList fetches the list of available course selection sessions
func getSessionList(cookies []*http.Cookie) ([]CourseSession, error) {
	req, err := http.NewRequest("GET", "https://jw.educationgroup.cn/ytkjxy_jsxsd/xsxk/xklc_list", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Host", "jw.educationgroup.cn")

	// Add cookies to request
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get session list with status code: %d", resp.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	html := string(body)

	// Extract all table rows with a more specific pattern for the table format
	var sessions []CourseSession
	sessionMap := make(map[string]CourseSession) // Use a map to avoid duplicates

	// Step 1: Try to find the specific table by ID or class
	tablePattern := regexp.MustCompile(`<table[^>]*(?:id=["']?tbKxkc["']?|class=["']?Nsb_r_list Nsb_table["']?)[^>]*>(?s:.*?)</table>`)
	tableMatch := tablePattern.FindString(html)

	if tableMatch != "" {
		fmt.Println("找到选课表格")

		// Step 2: Extract all rows from the table
		rowPattern := regexp.MustCompile(`<tr>[\s\S]*?</tr>`)
		allRows := rowPattern.FindAllString(tableMatch, -1)

		// Filter out header rows
		var dataRows []string
		for _, row := range allRows {
			// Skip rows that contain header cells or have the specific style attribute
			if !strings.Contains(row, "<th") && !strings.Contains(row, "background-color:#D1E4F8") {
				dataRows = append(dataRows, row)
			}
		}

		if len(dataRows) > 0 {
			fmt.Printf("找到 %d 行选课会话信息\n", len(dataRows))

			for _, rowHTML := range dataRows {
				// Remove HTML comments to avoid confusion
				rowWithoutComments := removeHTMLComments(rowHTML)

				// Extract the text from each cell
				cellPattern := regexp.MustCompile(`<td[^>]*>([\s\S]*?)</td>`)
				cellMatches := cellPattern.FindAllStringSubmatch(rowWithoutComments, -1)

				if len(cellMatches) >= 3 {
					// First three cells should contain term, name, and time
					term := strings.TrimSpace(cellMatches[0][1])
					name := strings.TrimSpace(cellMatches[1][1])

					// Try to extract time from the third cell
					timeStr := ""
					if len(cellMatches) >= 3 {
						timeStr = strings.TrimSpace(cellMatches[2][1])
					}

					// If time is empty, try to find it in any cell by looking for time patterns
					if timeStr == "" {
						timePattern := regexp.MustCompile(`\d{4}-\d{2}-\d{2}.*?~.*?\d{4}-\d{2}-\d{2}`)
						for _, cell := range cellMatches {
							if timeMatch := timePattern.FindString(cell[1]); timeMatch != "" {
								timeStr = timeMatch
								break
							}
						}
					}

					// If we still don't have time, look for the cell containing a date pattern
					if timeStr == "" {
						datePattern := regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)
						for _, cell := range cellMatches {
							if dateMatch := datePattern.FindString(cell[1]); dateMatch != "" {
								timeStr = strings.TrimSpace(cell[1])
								break
							}
						}
					}

					// Extract the last cell for operation links
					operationCell := cellMatches[len(cellMatches)-1][1]

					// Extract links from the operation cell
					linkPattern := regexp.MustCompile(`<a[^>]*href=["']([^"']*)["'][^>]*>([\s\S]*?)</a>`)
					linkMatches := linkPattern.FindAllStringSubmatch(operationCell, -1)

					for _, linkMatch := range linkMatches {
						href := linkMatch[1]
						linkText := strings.TrimSpace(linkMatch[2])

						// Clean HTML tags from extracted text
						cleanText := func(s string) string {
							// Remove HTML tags
							noTags := regexp.MustCompile(`<[^>]*>`).ReplaceAllString(s, "")
							return strings.TrimSpace(noTags)
						}

						term = cleanText(term)
						name = cleanText(name)
						timeStr = cleanText(timeStr)

						fmt.Printf("从表格提取: 学期=%s, 名称=%s, 时间=%s, 操作=%s\n",
							term, name, timeStr, linkText)

						// Convert xklc_view URLs to yxxsxk_index URLs if needed
						sessionURL := href

						// Extract all parameters from the URL without assuming specific names
						if strings.Contains(sessionURL, "xklc_view") {
							// Split the URL to get the path and parameters
							urlParts := strings.SplitN(sessionURL, "?", 2)
							if len(urlParts) == 2 {
								basePath := strings.Replace(urlParts[0], "xklc_view", "yxxsxk_index", 1)
								sessionURL = basePath + "?" + urlParts[1]
								fmt.Printf("转换URL: %s => %s\n", href, sessionURL)
							}
						} else if strings.Contains(sessionURL, "xsxk_index") {
							// Also convert any xsxk_index URLs to yxxsxk_index
							urlParts := strings.SplitN(sessionURL, "?", 2)
							if len(urlParts) == 2 {
								basePath := strings.Replace(urlParts[0], "xsxk_index", "yxxsxk_index", 1)
								sessionURL = basePath + "?" + urlParts[1]
								fmt.Printf("转换URL: %s => %s\n", href, sessionURL)
							}
						}

						// Use the full URL as the key for deduplication (don't rely on specific parameters)
						sessionKey := sessionURL

						sessionMap[sessionKey] = CourseSession{
							Term: term,
							Name: name,
							Time: timeStr,
							URL:  sessionURL,
						}
					}
				}
			}
		}
	}

	// If we couldn't extract from the table, try other approaches
	if len(sessionMap) == 0 {
		fmt.Println("未从表格中提取到选课会话，尝试通用提取方法...")

		// Look for any a tags with href containing xklc_view or xsxk_index
		linkPattern := regexp.MustCompile(`<a[^>]*href=["']([^"']*)["'][^>]*>([\s\S]*?)</a>`)
		linkMatches := linkPattern.FindAllStringSubmatch(html, -1)

		var selectionLinks [][]string
		for _, match := range linkMatches {
			href := match[1]
			linkText := match[2]

			if (strings.Contains(href, "xsxk") || strings.Contains(href, "xklc")) &&
				(strings.Contains(linkText, "选课") || strings.Contains(linkText, "进入")) {
				selectionLinks = append(selectionLinks, match)
			}
		}

		fmt.Printf("找到 %d 个可能的选课链接\n", len(selectionLinks))

		for _, match := range selectionLinks {
			href := match[1]
			linkText := strings.TrimSpace(match[2])

			// Try to find the containing table row to extract metadata
			// Create a regex pattern that will match a <tr> containing this href
			escapedHref := regexp.QuoteMeta(href)
			rowPattern := regexp.MustCompile(`<tr>[\s\S]*?` + escapedHref + `[\s\S]*?</tr>`)
			rowMatch := rowPattern.FindString(html)

			// Default values
			term := "当前学期"
			name := linkText
			timeStr := "当前时间"

			if rowMatch != "" {
				// Remove HTML comments to avoid confusion
				rowMatch = removeHTMLComments(rowMatch)

				// Extract cells from the row
				cellPattern := regexp.MustCompile(`<td[^>]*>([\s\S]*?)</td>`)
				cellMatches := cellPattern.FindAllStringSubmatch(rowMatch, -1)

				if len(cellMatches) >= 3 {
					// Try to extract text content without HTML tags
					extractText := func(html string) string {
						// Remove HTML tags
						noTags := regexp.MustCompile(`<[^>]*>`).ReplaceAllString(html, "")
						// Trim whitespace
						return strings.TrimSpace(noTags)
					}

					// First three cells should contain term, name, and time
					term = extractText(cellMatches[0][1])
					name = extractText(cellMatches[1][1])

					// Try to extract time from the third cell
					if len(cellMatches) >= 3 {
						timeStr = extractText(cellMatches[2][1])
					}

					// If time is empty, try to find it in any cell by looking for time patterns
					if timeStr == "" || timeStr == "当前时间" {
						timePattern := regexp.MustCompile(`\d{4}-\d{2}-\d{2}.*?~.*?\d{4}-\d{2}-\d{2}`)
						for _, cell := range cellMatches {
							if timeMatch := timePattern.FindString(cell[1]); timeMatch != "" {
								timeStr = extractText(timeMatch)
								break
							}
						}
					}

					// If we still don't have time, look for the cell containing a date pattern
					if timeStr == "" || timeStr == "当前时间" {
						datePattern := regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)
						for _, cell := range cellMatches {
							if dateMatch := datePattern.FindString(cell[1]); dateMatch != "" {
								timeStr = extractText(cell[1])
								break
							}
						}
					}
				}
			}

			// Convert xklc_view URLs to yxxsxk_index URLs if needed
			sessionURL := href

			// Extract all parameters from the URL without assuming specific names
			if strings.Contains(sessionURL, "xklc_view") {
				// Split the URL to get the path and parameters
				urlParts := strings.SplitN(sessionURL, "?", 2)
				if len(urlParts) == 2 {
					basePath := strings.Replace(urlParts[0], "xklc_view", "yxxsxk_index", 1)
					sessionURL = basePath + "?" + urlParts[1]
					fmt.Printf("转换URL: %s => %s\n", href, sessionURL)
				}
			} else if strings.Contains(sessionURL, "xsxk_index") {
				// Also convert any xsxk_index URLs to yxxsxk_index
				urlParts := strings.SplitN(sessionURL, "?", 2)
				if len(urlParts) == 2 {
					basePath := strings.Replace(urlParts[0], "xsxk_index", "yxxsxk_index", 1)
					sessionURL = basePath + "?" + urlParts[1]
					fmt.Printf("转换URL: %s => %s\n", href, sessionURL)
				}
			}

			// Use the full URL as the key for deduplication
			sessionKey := sessionURL

			// Extract all parameters for logging purposes only
			paramPattern := regexp.MustCompile(`([a-zA-Z0-9_]+)=([^&]+)`)
			paramMatches := paramPattern.FindAllStringSubmatch(sessionURL, -1)
			for _, paramMatch := range paramMatches {
				if len(paramMatch) >= 3 {
					paramName := paramMatch[1]
					paramValue := paramMatch[2]
					fmt.Printf("提取到参数: %s=%s\n", paramName, paramValue)
				}
			}

			sessionMap[sessionKey] = CourseSession{
				Term: term,
				Name: name,
				Time: timeStr,
				URL:  sessionURL,
			}
		}
	}

	// Convert map to slice
	for _, session := range sessionMap {
		sessions = append(sessions, session)
	}

	if len(sessions) > 0 {
		fmt.Printf("\n找到 %d 个唯一的选课会话\n", len(sessions))
		for i, session := range sessions {
			fmt.Printf("会话 %d: %s - %s - %s - %s\n", i+1, session.Term, session.Name, session.Time, session.URL)
		}
		return sessions, nil
	}

	// If we still couldn't find any sessions, return an error
	return nil, fmt.Errorf("无法从响应中提取选课会话信息")
}

// removeHTMLComments removes HTML comments from a string
func removeHTMLComments(html string) string {
	commentPattern := regexp.MustCompile(`<!--[\s\S]*?-->`)
	return commentPattern.ReplaceAllString(html, "")
}
