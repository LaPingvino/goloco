// Generates pkg/sim/testdata/calcdays.txt from Date.cpp's calcDays / getMonthTotalDay / isLeapYear (copied by sed, see tick 254).
#include <cstdio>
#include <cstdint>
#include <utility>
enum class MonthId : uint8_t { january, february, march, april, may, june, july, august, september, october, november, december };
struct Date { int32_t day=0; MonthId month; int32_t year=0; int32_t dayOfYear=0; Date(int32_t y, MonthId m, int32_t d):day(d),month(m),year(y){} };
    bool isLeapYear(const int year)
    {
        return year % 4 == 0;
    }
uint8_t getMonthTotalDay(uint16_t year, MonthId month);
    uint32_t calcDays(Date date)
    {
        constexpr auto kBaseYear = 1800;
        constexpr auto kDaysInYear = 365;

        // adds years (365 for each year + 1 for leap years)
        auto yearDiff = date.year - kBaseYear;
        uint32_t dayCount = (yearDiff * kDaysInYear) + (yearDiff / 4);

        // add months
        for (int month = 0; month < static_cast<uint8_t>(date.month); ++month)
        {
            dayCount += getMonthTotalDay(date.year, static_cast<MonthId>(month));
        }

        // add days
        dayCount += date.day;

        return dayCount;
    }
    uint8_t getMonthTotalDay(uint16_t year, MonthId month)
    {
        static constexpr std::pair<MonthId, uint8_t> month_table[] = {
            { MonthId::january, 31 },
            { MonthId::february, 28 },
            { MonthId::march, 31 },
            { MonthId::april, 30 },
            { MonthId::may, 31 },
            { MonthId::june, 30 },
            { MonthId::july, 31 },
            { MonthId::august, 31 },
            { MonthId::september, 30 },
            { MonthId::october, 31 },
            { MonthId::november, 30 },
            { MonthId::december, 31 },
        };
        bool extraDay = month == MonthId::february && isLeapYear(year);
        return month_table[(uint8_t)month].second + extraDay;
    }
int main(){ for(int y=1800;y<2101;y+=7) for(int m=0;m<12;m++) for(int d=1;d<=28;d+=9) printf("%d %d %d %u\n",y,m,d,calcDays(Date(y,(MonthId)m,d))); }
