import numpy as np
import pandas as pd
import time
from functools import reduce

# stocks: pandas.Dataframe
# stocks, stocks_amount
class EngineChunked():
    def __init__(self, logger):
        self.logger = logger 
        
    def create_portfolio(self, dfs, stocks_amount, elite, chunk_number, iterations_in_chunk):
        stocks = reduce(lambda left,right: pd.merge(left,right,on='Date'), dfs)

        def hist_return(months):
            ''' It calculates Stock returns for various months and returns a dataframe.
                Input: Months in the form of a list.
                Output: Historical returns in the form of a DataFrame. '''
            idx=[]
            df=pd.DataFrame()
            for mon in months:
                temp=(stocks.iloc[0,1:] - stocks.iloc[mon,1:])/(stocks.iloc[mon,1:])
                idx.append(str(mon)+'_mon_return')
                df=pd.concat([df, temp.to_frame().T], ignore_index=True)
            df.index=idx
            return df    


        def gen_mc_grid(rows, cols, n, N):  # , xfname): generate monte carlo wind farm layout grids
            np.random.seed(seed=int(time.time()))  # init random seed
            layouts = np.zeros((n, rows * cols), dtype=np.int32)  # one row is a layout
            positionX = np.random.randint(0, cols, size=(N * n * 2))
            positionY = np.random.randint(0, rows, size=(N * n * 2))
            ind_rows = 0  # index of layouts from 0 to n-1
            ind_pos = 0  # index of positionX, positionY from 0 to N*n*2-1
            while ind_rows < n:
                layouts[ind_rows, positionX[ind_pos] + positionY[ind_pos] * cols] = 1
                if np.sum(layouts[ind_rows, :]) == N:
                    ind_rows += 1
                ind_pos += 1
                if ind_pos >= N * n * 2:
                    self.logger.info("Not enough positions")
                    break
            return layouts

        def gen_mc_grid_with_NA_loc(rows, cols, n, N,NA_loc):  # , xfname): generate monte carlo wind farm layout grids
            np.random.seed(seed=int(time.time()))  # init random seed
            layouts = np.zeros((n, rows * cols), dtype=np.int32)  # one row is a layout, NA loc is 0

            layouts_NA= np.zeros((n, rows * cols), dtype=np.int32)  # one row is a layout, NA loc is 2
            for i in NA_loc:
                layouts_NA[:,i-1]=2

            positionX = np.random.randint(0, cols, size=(N * n * 2))
            positionY = np.random.randint(0, rows, size=(N * n * 2))
            ind_rows = 0  # index of layouts from 0 to n-1
            ind_pos = 0  # index of positionX, positionY from 0 to N*n*2-1
            N_count=0
            while ind_rows < n:
                cur_state=layouts_NA[ind_rows, positionX[ind_pos] + positionY[ind_pos] * cols]
                if cur_state!=1 and cur_state!=2:
                    layouts[ind_rows, positionX[ind_pos] + positionY[ind_pos] * cols]=1
                    layouts_NA[ind_rows, positionX[ind_pos] + positionY[ind_pos] * cols] = 1
                    N_count+=1
                    if np.sum(layouts[ind_rows, :]) == N:
                        ind_rows += 1
                        N_count=0
                ind_pos += 1
                if ind_pos >= N * n * 2:
                    self.logger.info("Not enough positions")
                    break
            return layouts,layouts_NA

        def chromosome(n):
            ''' Generates set of random numbers whose sum is equal to 1
                Input: Number of stocks.
                Output: Array of random numbers'''
            ch = np.random.rand(n)
            return ch/sum(ch)


        hist_stock_returns=hist_return([3,6,12,24,36])


        n= stocks_amount #6 # Number of stocks = 6
        pop_size=100 # initial population = 100

        population = np.array([chromosome(n) for _ in range(pop_size)])


        # Calculate covariance of historical returns
        cov_hist_return=hist_stock_returns.cov()

        # For ease of calculations make covariance of same variable as zero.
        for i in range(6):
            cov_hist_return.iloc[i][i]=0


        # Calculate the mean of historical returns
        mean_hist_return=hist_stock_returns.mean()


        # Calculate Standard deviation of historical returns:
        sd_hist_return=hist_stock_returns.std()


        # Calculate Expected returns of portfolio.
        # Mean portfolio return = Mean Return * Fractions of Total Capital (Chromosome).
        def mean_portfolio_return(child):
            return np.sum(np.multiply(child,mean_hist_return))

        # Standard deviation of portfolio return = (chromosome * Standard deviation)**2 + Covariance * Respective weights in chromosome.

        # Calculate portfolio variance.
        def var_portfolio_return(child):
            part_1 = np.sum(np.multiply(child,sd_hist_return)**2)
            temp_lst=[]
            for i in range(6):
                for j in range(6):
                    temp=cov_hist_return.iloc[i][j] * child[i] * child[j]
                    temp_lst.append(temp)
            part_2=np.sum(temp_lst)
            return part_1+part_2


        # S = (µ − r)/σ
        # Here µ is the return of the portfolio over a specified period or Mean portfolio return, 
        #      r is the risk-free rate over the same period and 
        #      σ is the standard deviation of the returns over the specified period or Standard deviation of portfolio return.
        rf= 0.0697
        def fitness_fuction(child):
            ''' This will return the Sharpe ratio for a particular portfolio.
                Input: A child/chromosome (1D Array)
                Output: Sharpe Ratio value (Scalar)'''
            return (mean_portfolio_return(child)-rf)/np.sqrt(var_portfolio_return(child))


        def Select_elite_population(population, frac=0.3):
            ''' Select elite population from the total population based on fitness function values.
                Input: Population and fraction of population to be considered as elite.
                Output: Elite population.'''
            population = sorted(population,key = lambda x: fitness_fuction(x),reverse=True)
            percentage_elite_idx = int(np.floor(len(population)* frac))
            return population[:percentage_elite_idx]



        def mutation(parent):
            ''' Randomy choosen elements of a chromosome are swapped
                Input: Parent
                Output: Offspring (1D Array)'''
            child=parent.copy()
            n=np.random.choice(range(6),2)
            while (n[0]==n[1]):
                n=np.random.choice(range(6),2)
            child[n[0]],child[n[1]]=child[n[1]],child[n[0]]
            return child/sum(child)


        def Heuristic_crossover(parent1,parent2):
            ''' The oﬀsprings are created according to the equation:
                    Off_spring A = Best Parent  + β ∗ ( Best Parent − Worst Parent)
                    Off_spring B = Worst Parent - β ∗ ( Best Parent − Worst Parent)
                        Where β is a random number between 0 and 1.
                Input: 2 Parents
                Output: 2 Children (1d Array)'''
            ff1=fitness_fuction(parent1)
            ff2=fitness_fuction(parent2)
            diff=parent1 - parent2
            beta=np.random.rand()

            child1=np.abs(parent1 + beta * diff)
            child2=np.abs(parent2 - beta * diff)
            if not ff1>ff2:
                child2, child1 = child1, child2  

            return child1/sum(child1), child2/sum(child2)

        def Arithmetic_crossover(parent1,parent2):
            ''' The oﬀsprings are created according to the equation:
                    Off spring A = α ∗ Parent1 + (1 −α) ∗ Parent2
                    Off spring B = (1 −α) ∗ Parent1 + α ∗ Parent2
                    
                        Where α is a random number between 0 and 1.
                Input: 2 Parents
                Output: 2 Children (1d Array)'''
            alpha = np.random.rand()
            child1 = alpha * parent1 + (1-alpha) * parent2
            child2 = (1-alpha) * parent1 + alpha * parent2
            return child1/sum(child1), child2/sum(child2)

        def Geometric_crossover(parent1, parent2):
            ''' The oﬀsprings are created according to the equation:
                    Off spring A = Parent1**α *Parent2**(1 −α)
                    Off spring B = Parent1**(1 −α) *Parent2**α
                    
                        Where α is a random number between 0 and 1.
                Input: 2 Parents
                Output: 2 Children (1d Array)'''
            alpha = np.random.rand()
            child1 = (np.abs(parent1)) ** (alpha) * (np.abs(parent2)) ** (1-alpha)
            child2 = (np.abs(parent1)) ** (1-alpha) * (np.abs(parent2)) ** (alpha)
            return child1/sum(child1), child2/sum(child2)


        def next_generation(pop_size,elite,crossover=Heuristic_crossover):
            ''' Generates new population from elite population with mutation probability as 0.4 and crossover as 0.6. 
                Over the final stages, mutation probability is decreased to 0.1.
                Input: Population Size and elite population.
                Output: Next generation population (2D Array).'''
            new_population=[]
            elite_range=range(len(elite))
            while len(new_population) < pop_size:
                if len(new_population) > 2*pop_size/3: # In the final stages mutation frequency is decreased.
                    mutate_or_crossover = np.random.choice([0, 1], p=[0.9, 0.1])
                else:
                    mutate_or_crossover = np.random.choice([0, 1], p=[0.4, 0.6])
                if mutate_or_crossover:
                    indx=np.random.choice(elite_range)
                    new_population.append(mutation(elite[indx]))
                else:
                    p1_idx,p2_idx=np.random.choice(elite_range,2)
                    c1,c2=crossover(elite[p1_idx],elite[p2_idx])
                    chk=0
                    new_population.extend([c1,c2])
            return new_population


        # With Heuristic_crossover:
        n=6 # Number of stocks = 6
        pop_size=100 # initial population = 100

        if chunk_number == 0:
            # Initial population
            population = np.array([chromosome(n) for _ in range(pop_size)])

            # Get initial elite population
            elite = Select_elite_population(population)
            # self.logger.info(elite[0], fitness_fuction(elite[0]))


        iteration = chunk_number * iterations_in_chunk

        iteration=0
        Expected_returns=0
        Expected_risk=1
        #  (Expected_returns < 0.30 and Expected_risk > 0.0005) or
        while iteration <= (chunk_number+1) * iterations_in_chunk:
            
            # self.logger.info('Iteration:',iteration)
            population = next_generation(100,elite)
            elite = Select_elite_population(population)
            Expected_returns=mean_portfolio_return(elite[0])
            Expected_risk=var_portfolio_return(elite[0])
            fitness = fitness_fuction(elite[0])
            # self.logger.info(elite[0], fitness_fuction(elite[0]))

            iteration+=1


        # self.logger.info('Portfolio of stocks after all the iterations:\n')
        # [self.logger.info(hist_stock_returns.columns[i],':',elite[0][i]) for i in list(range(6))]

        # self.logger.info('\nExpected returns of {} with risk of {}\n'.format(Expected_returns,Expected_risk))
        return Expected_returns, Expected_risk, hist_stock_returns, elite
